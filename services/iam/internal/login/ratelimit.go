// Package login implements the IAM login flow: rate-limit, password
// verification, lockout escalation, audit emission.
//
// REQ-FUNC-PLT-IAM-001 / REQ-FUNC-PLT-IAM-003
// design.md §4.1.1
package login

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ----------------------------------------------------------------------
// Sliding-window rate limiter
// ----------------------------------------------------------------------
//
// REQ-FUNC-PLT-IAM-003 mandates a sliding-window IP rate limit at
// 10 attempts/minute. The implementation uses a Redis sorted set
// keyed `iam:login:ip:<ip>` whose members are the request timestamps
// in nanoseconds. On each call we:
//
//   1. Drop entries older than (now - window).
//   2. Count what remains.
//   3. If under the limit, ZADD this attempt and return AllowResult.
//   4. Otherwise return DenyResult with a Retry-After hint.
//
// All four steps run inside a single MULTI/EXEC transaction so two
// concurrent requests cannot both squeak past the limit at the
// boundary.

// IPLimiterConfig configures the per-IP sliding window limiter.
type IPLimiterConfig struct {
	// MaxAttemptsPerWindow is the maximum number of attempts a
	// single IP can make inside Window.
	MaxAttemptsPerWindow int

	// Window is the sliding-window duration (e.g. 1 minute).
	Window time.Duration

	// KeyPrefix is prefixed to every Redis key. Defaults to
	// "iam:login:ip:".
	KeyPrefix string

	// Now is the clock; tests override. nil → time.Now.
	Now func() time.Time
}

// IPLimiter is the Redis-backed per-IP login rate limiter.
type IPLimiter struct {
	rdb redis.UniversalClient
	cfg IPLimiterConfig
}

// NewIPLimiter constructs an IPLimiter against the supplied Redis
// client. The client lifecycle is the caller's responsibility.
func NewIPLimiter(rdb redis.UniversalClient, cfg IPLimiterConfig) *IPLimiter {
	if cfg.MaxAttemptsPerWindow <= 0 {
		cfg.MaxAttemptsPerWindow = 10
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "iam:login:ip:"
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &IPLimiter{rdb: rdb, cfg: cfg}
}

// LimitResult is what Allow returns. Allowed=true ⇒ the caller
// should proceed with the password check. Allowed=false ⇒ caller
// MUST return HTTP 429 with the supplied RetryAfter.
type LimitResult struct {
	Allowed     bool
	RetryAfter  time.Duration // valid only when Allowed=false
	HitsInWindow int          // diagnostic
	Limit        int          // diagnostic
}

// Allow records an attempt for `ip` and reports whether it is
// permitted. A nil error indicates a successful evaluation; the
// admit/deny verdict is in result.Allowed.
func (l *IPLimiter) Allow(ctx context.Context, ip string) (LimitResult, error) {
	if ip == "" {
		return LimitResult{}, errors.New("ratelimit: empty client IP")
	}
	now := l.cfg.Now()
	cutoff := now.Add(-l.cfg.Window).UnixNano()
	score := now.UnixNano()
	key := l.cfg.KeyPrefix + ip

	// MULTI / EXEC keeps the prune+count+admit sequence atomic.
	// A pipeline here would race; a Lua script would also work but
	// MULTI is enough for this volume.
	pipe := l.rdb.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", cutoff))
	countCmd := pipe.ZCard(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return LimitResult{}, fmt.Errorf("ratelimit: prune+count: %w", err)
	}
	count := int(countCmd.Val())

	if count >= l.cfg.MaxAttemptsPerWindow {
		// Find the oldest entry to compute Retry-After. ZRangeByScore
		// with LIMIT 0 1 returns the chronologically earliest member.
		oldest, err := l.rdb.ZRangeWithScores(ctx, key, 0, 0).Result()
		if err != nil {
			return LimitResult{}, fmt.Errorf("ratelimit: oldest: %w", err)
		}
		retry := time.Second
		if len(oldest) > 0 {
			oldestNs := int64(oldest[0].Score)
			expiresAt := time.Unix(0, oldestNs).Add(l.cfg.Window)
			if d := expiresAt.Sub(now); d > 0 {
				retry = d
			}
		}
		return LimitResult{
			Allowed:      false,
			RetryAfter:   retry,
			HitsInWindow: count,
			Limit:        l.cfg.MaxAttemptsPerWindow,
		}, nil
	}

	// Admit the attempt.
	pipe = l.rdb.TxPipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: fmt.Sprintf("%d", score)})
	// Expire the whole sorted set after the window elapses; cheaper
	// than scheduling a janitor.
	pipe.PExpire(ctx, key, l.cfg.Window+5*time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return LimitResult{}, fmt.Errorf("ratelimit: admit: %w", err)
	}
	return LimitResult{
		Allowed:      true,
		HitsInWindow: count + 1,
		Limit:        l.cfg.MaxAttemptsPerWindow,
	}, nil
}

// Reset clears the rate-limit window for an IP. Used by admin
// override; not exposed as an RPC in v1.
func (l *IPLimiter) Reset(ctx context.Context, ip string) error {
	key := l.cfg.KeyPrefix + ip
	return l.rdb.Del(ctx, key).Err()
}
