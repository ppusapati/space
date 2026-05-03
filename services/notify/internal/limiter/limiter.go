// Package limiter implements the per-user SMS rate cap.
//
// → REQ-FUNC-PLT-NOTIFY-002: 5 SMS / hour / user.
//
// Backed by a Redis sliding-window using sorted sets — same shape
// the IAM login rate limiter uses (ratelimit.go in services/iam).
// The window key is `notify:sms:rate:{user_id}`; entries are
// keyed by send-time epoch milliseconds.

package limiter

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Defaults per REQ-FUNC-PLT-NOTIFY-002.
const (
	DefaultLimit  = 5
	DefaultWindow = time.Hour
)

// Result is the outcome of an Allow call.
type Result struct {
	Allowed       bool
	HitsInWindow  int
	Limit         int
	RetryAfter    time.Duration
}

// Config configures the limiter.
type Config struct {
	Limit  int
	Window time.Duration
	Now    func() time.Time
}

// SMSLimiter caps SMS sends per user.
type SMSLimiter struct {
	rdb *redis.Client
	cfg Config
}

// NewSMSLimiter wraps a Redis client.
func NewSMSLimiter(rdb *redis.Client, cfg Config) (*SMSLimiter, error) {
	if rdb == nil {
		return nil, errors.New("limiter: nil redis client")
	}
	if cfg.Limit <= 0 {
		cfg.Limit = DefaultLimit
	}
	if cfg.Window <= 0 {
		cfg.Window = DefaultWindow
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &SMSLimiter{rdb: rdb, cfg: cfg}, nil
}

// Allow checks (and on success, records) one SMS send for userID.
// Returns Allowed=false when the user has hit Limit in the past
// Window; the RetryAfter field is the delay until the oldest
// entry rolls out of the window.
func (l *SMSLimiter) Allow(ctx context.Context, userID string) (Result, error) {
	if userID == "" {
		return Result{}, errors.New("limiter: empty user_id")
	}
	now := l.cfg.Now().UTC()
	cutoff := now.Add(-l.cfg.Window)
	key := fmt.Sprintf("notify:sms:rate:{%s}", userID)
	score := strconv.FormatInt(now.UnixMilli(), 10)

	pipe := l.rdb.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(cutoff.UnixMilli(), 10))
	cardCmd := pipe.ZCard(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return Result{}, fmt.Errorf("limiter: prune: %w", err)
	}
	currentCount := int(cardCmd.Val())

	if currentCount >= l.cfg.Limit {
		// Compute RetryAfter from the oldest remaining entry.
		oldest, err := l.rdb.ZRangeWithScores(ctx, key, 0, 0).Result()
		if err != nil || len(oldest) == 0 {
			return Result{
				Allowed:      false,
				HitsInWindow: currentCount,
				Limit:        l.cfg.Limit,
				RetryAfter:   l.cfg.Window,
			}, nil
		}
		retry := time.Duration(int64(oldest[0].Score)-cutoff.UnixMilli()) * time.Millisecond
		if retry < 0 {
			retry = 0
		}
		return Result{
			Allowed:      false,
			HitsInWindow: currentCount,
			Limit:        l.cfg.Limit,
			RetryAfter:   retry,
		}, nil
	}

	// Record the send + extend the key TTL so it auto-expires when
	// the window goes idle.
	pipe = l.rdb.TxPipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixMilli()), Member: score})
	pipe.Expire(ctx, key, l.cfg.Window+time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return Result{}, fmt.Errorf("limiter: record: %w", err)
	}
	return Result{
		Allowed:      true,
		HitsInWindow: currentCount + 1,
		Limit:        l.cfg.Limit,
	}, nil
}
