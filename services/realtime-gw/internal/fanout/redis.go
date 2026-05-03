// Package fanout implements the chetana realtime fan-out fabric.
//
// Two pieces:
//
//   • redis.go — local fan-out across all connections subscribed
//     to a topic. The realtime gateway is horizontally scalable
//     because every replica subscribes to the same Redis Pub/Sub
//     channel; messages produced by ANY replica reach EVERY
//     replica via Redis, then each replica's local
//     ConnectionRegistry walks its own connection set.
//
//   • kafka.go — bridges chetana's domain Kafka topics
//     (telemetry.params, pass.state, alert.*, command.state,
//     notify.inapp.v1) into the Redis Pub/Sub channels the WS
//     layer subscribes to.

package fanout

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// Channel returns the Redis Pub/Sub channel name for `topic`.
// Convention: `chetana:rt:<topic>` so the namespace is clearly
// the realtime gateway's.
func Channel(topic string) string { return "chetana:rt:" + topic }

// Receiver is invoked per message arriving on a topic.
type Receiver func(ctx context.Context, topic string, payload []byte)

// RedisFanout is a thin wrapper over redis.Client's Pub/Sub.
type RedisFanout struct {
	rdb *redis.Client

	mu       sync.Mutex
	subs     map[string]*redis.PubSub // topic → subscription
}

// NewRedisFanout wraps a redis client.
func NewRedisFanout(rdb *redis.Client) (*RedisFanout, error) {
	if rdb == nil {
		return nil, errors.New("fanout: nil redis client")
	}
	return &RedisFanout{rdb: rdb, subs: map[string]*redis.PubSub{}}, nil
}

// Publish sends `payload` to every replica subscribed to `topic`.
// Best-effort: the call returns the underlying redis error but
// does NOT block on consumer ack — Pub/Sub is fire-and-forget.
func (f *RedisFanout) Publish(ctx context.Context, topic string, payload []byte) error {
	if topic == "" {
		return errors.New("fanout: empty topic")
	}
	if err := f.rdb.Publish(ctx, Channel(topic), payload).Err(); err != nil {
		return fmt.Errorf("fanout: publish: %w", err)
	}
	return nil
}

// Subscribe registers `recv` for messages on `topic`. The first
// subscriber for a topic opens the underlying Redis subscription
// + spawns a goroutine that pumps messages into `recv`. Returns
// a cancel func that unsubscribes when called.
func (f *RedisFanout) Subscribe(ctx context.Context, topic string, recv Receiver) (cancel func(), err error) {
	if topic == "" {
		return nil, errors.New("fanout: empty topic")
	}
	if recv == nil {
		return nil, errors.New("fanout: nil receiver")
	}
	f.mu.Lock()
	sub, exists := f.subs[topic]
	if !exists {
		sub = f.rdb.Subscribe(ctx, Channel(topic))
		f.subs[topic] = sub
	}
	f.mu.Unlock()

	// Confirm the sub is healthy.
	if _, err := sub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("fanout: subscribe: %w", err)
	}

	pumpCtx, pumpCancel := context.WithCancel(ctx)
	go func() {
		ch := sub.Channel()
		for {
			select {
			case <-pumpCtx.Done():
				return
			case m, ok := <-ch:
				if !ok {
					return
				}
				recv(pumpCtx, topic, []byte(m.Payload))
			}
		}
	}()
	return pumpCancel, nil
}

// Close stops every active subscription. Called by cmd/realtime-gw
// at shutdown.
func (f *RedisFanout) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	var firstErr error
	for _, s := range f.subs {
		if err := s.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	f.subs = map[string]*redis.PubSub{}
	return firstErr
}
