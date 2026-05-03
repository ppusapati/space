// kafka.go — bridges Kafka topic events into the Redis Pub/Sub
// fan-out so every gateway replica picks them up.
//
// One consumer group per gateway deployment; each Kafka topic is
// re-published onto the corresponding Redis channel via
// `Channel(topic)`. The WS layer's local subscriber walks its
// connection registry and pushes to each subscribed connection's
// backpressure buffer.

package fanout

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

// KafkaBridge consumes one or more Kafka topics + republishes the
// payloads onto Redis Pub/Sub for the gateway's local readers.
type KafkaBridge struct {
	consumer sarama.ConsumerGroup
	fanout   *RedisFanout
	topics   []string
	groupID  string
}

// NewKafkaBridge wires a sarama ConsumerGroup + RedisFanout.
func NewKafkaBridge(consumer sarama.ConsumerGroup, fanout *RedisFanout, topics []string, groupID string) (*KafkaBridge, error) {
	if consumer == nil {
		return nil, errors.New("fanout: nil consumer")
	}
	if fanout == nil {
		return nil, errors.New("fanout: nil redis fanout")
	}
	if len(topics) == 0 {
		return nil, errors.New("fanout: at least one topic required")
	}
	if groupID == "" {
		groupID = "realtime-gw"
	}
	return &KafkaBridge{consumer: consumer, fanout: fanout, topics: topics, groupID: groupID}, nil
}

// Run blocks until ctx is cancelled. Each message → Redis
// Publish on the corresponding chetana channel.
func (b *KafkaBridge) Run(ctx context.Context) error {
	handler := &bridgeHandler{fanout: b.fanout}
	for {
		if err := b.consumer.Consume(ctx, b.topics, handler); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) || errors.Is(err, context.Canceled) {
				return ctx.Err()
			}
			return fmt.Errorf("fanout: consumer: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// bridgeHandler implements sarama.ConsumerGroupHandler.
type bridgeHandler struct {
	fanout *RedisFanout
	wg     sync.WaitGroup
}

// Setup is called when a session begins.
func (h *bridgeHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called when the session ends.
func (h *bridgeHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	h.wg.Wait()
	return nil
}

// ConsumeClaim drains messages from one partition.
func (h *bridgeHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// Republish onto Redis. Best-effort: a Pub/Sub failure
		// does NOT NACK the Kafka message — the gateway treats
		// realtime fan-out as best-effort fan-out.
		_ = h.fanout.Publish(session.Context(), msg.Topic, msg.Value)
		session.MarkMessage(msg, "")
	}
	return nil
}
