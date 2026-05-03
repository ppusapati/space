// Package lock implements the distributed per-job lock the
// scheduler uses to guarantee exactly-one runner per scheduled
// tick across N replicas (REQ-FUNC-CMN-006 acceptance #1).
//
// Algorithm:
//
//   • Acquire: SET NX EX — atomic + auto-expiring. The value is
//     a per-acquisition fencing token (UUID) that the holder
//     uses on Release so a delayed runner can't accidentally
//     unlock a lock that has elapsed + been re-acquired by
//     another runner ("Martin Kleppmann's fencing-token problem").
//
//   • Release: a small Lua CAS that compares the stored value
//     to the supplied token before deleting. Atomic on the Redis
//     side.
//
//   • Refresh: callers running long jobs MUST call Refresh
//     periodically (every TTL/2 is the chetana convention) so
//     the lock doesn't elapse mid-execution.

package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lock is a held distributed lock.
type Lock struct {
	rdb   *redis.Client
	key   string
	token string
	ttl   time.Duration
}

// Locker mints + manages locks.
type Locker struct {
	rdb *redis.Client
}

// NewLocker wraps a Redis client.
func NewLocker(rdb *redis.Client) (*Locker, error) {
	if rdb == nil {
		return nil, errors.New("lock: nil redis client")
	}
	return &Locker{rdb: rdb}, nil
}

// Acquire attempts to acquire `key` for `ttl`. Returns nil + nil
// when the lock is already held by someone else.
func (l *Locker) Acquire(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	if key == "" {
		return nil, errors.New("lock: empty key")
	}
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	token, err := newToken()
	if err != nil {
		return nil, fmt.Errorf("lock: token: %w", err)
	}
	ok, err := l.rdb.SetNX(ctx, fullKey(key), token, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("lock: setnx: %w", err)
	}
	if !ok {
		return nil, nil
	}
	return &Lock{rdb: l.rdb, key: fullKey(key), token: token, ttl: ttl}, nil
}

// releaseScript is the CAS-then-DEL Lua. Single round-trip.
var releaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
end
return 0
`)

// Release safely drops the lock if (and only if) we still hold
// it. Returns nil on success or when the lock was already
// elapsed; ErrLockLost when another holder owns it now.
func (lock *Lock) Release(ctx context.Context) error {
	res, err := releaseScript.Run(ctx, lock.rdb, []string{lock.key}, lock.token).Int()
	if err != nil {
		return fmt.Errorf("lock: release: %w", err)
	}
	if res == 0 {
		return ErrLockLost
	}
	return nil
}

// refreshScript: extend the TTL only when we still own the lock.
var refreshScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("EXPIRE", KEYS[1], ARGV[2])
end
return 0
`)

// Refresh extends the lock's TTL by `ttl` if we still hold it.
// Returns ErrLockLost when ownership has been lost.
func (lock *Lock) Refresh(ctx context.Context, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = lock.ttl
	}
	res, err := refreshScript.Run(ctx, lock.rdb, []string{lock.key}, lock.token, int(ttl.Seconds())).Int()
	if err != nil {
		return fmt.Errorf("lock: refresh: %w", err)
	}
	if res == 0 {
		return ErrLockLost
	}
	return nil
}

// Token returns the fencing token. Useful for callers that want
// to embed the token in their own request envelope (e.g. so a
// downstream RPC can reject stale work).
func (lock *Lock) Token() string { return lock.token }

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func fullKey(key string) string { return "scheduler:lock:" + key }

// ErrLockLost is returned by Release / Refresh when the supplied
// holder no longer owns the lock — typically because it elapsed
// + was re-acquired by another runner.
var ErrLockLost = errors.New("lock: lost ownership")
