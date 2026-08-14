// Package kafka is furumai's Kafka Stimulus/Observation adapter.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	segmentio "github.com/segmentio/kafka-go"
)

// Stimulus publishes to topics on a Kafka cluster. It is used from both
// given and when steps, since Stimulus adapters are shared between them.
type Stimulus struct {
	Brokers []string
}

// NewStimulus returns a Stimulus connecting to brokers.
func NewStimulus(brokers ...string) *Stimulus {
	return &Stimulus{Brokers: brokers}
}

// Publish writes a single message to topic.
func (s *Stimulus) Publish(topic string, key, value []byte, headers map[string]string) error {
	w := &segmentio.Writer{
		Addr:                   segmentio.TCP(s.Brokers...),
		Topic:                  topic,
		Balancer:               &segmentio.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer w.Close()

	msg := segmentio.Message{Key: key, Value: value}
	for k, v := range headers {
		msg.Headers = append(msg.Headers, segmentio.Header{Key: k, Value: []byte(v)})
	}

	return w.WriteMessages(context.Background(), msg)
}

// PublishJSON marshals v as JSON and publishes it as the message value.
func (s *Stimulus) PublishJSON(topic, key string, v any) error {
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.Publish(topic, []byte(key), value, nil)
}

// Message is the full-state Observation for a single Kafka message. Key,
// Value and Headers are typed any so a furumai.Matcher (Any, Regex, ...)
// can be substituted for any of them.
type Message struct {
	Key     any
	Value   any
	Headers any
}

// Subscription observes messages published to a topic after Listen was
// called. Kafka doesn't guarantee order across partitions, so wrap the
// expected []Message in furumai.AnyOrder when comparing.
type Subscription struct {
	reader *segmentio.Reader
}

// Listen starts consuming topic from this point forward. Call it before
// the when step expected to produce messages, so nothing published in the
// meantime is missed; then call Collect in then to observe what arrived.
func (s *Stimulus) Listen(topic string) (*Subscription, error) {
	reader := segmentio.NewReader(segmentio.ReaderConfig{
		Brokers:     s.Brokers,
		Topic:       topic,
		GroupID:     fmt.Sprintf("furumai-%d", time.Now().UnixNano()),
		StartOffset: segmentio.LastOffset,
	})
	return &Subscription{reader: reader}, nil
}

// Collect gathers every message received within window and returns them as
// the full-state Observation. An empty result means no message arrived.
func (sub *Subscription) Collect(window time.Duration) ([]Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), window)
	defer cancel()

	var messages []Message
	for {
		m, err := sub.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			return nil, err
		}
		messages = append(messages, Message{
			Key:     string(m.Key),
			Value:   string(m.Value),
			Headers: headersToMap(m.Headers),
		})
	}
	return messages, nil
}

// Close stops consuming and releases the underlying connection.
func (sub *Subscription) Close() error {
	return sub.reader.Close()
}

func headersToMap(headers []segmentio.Header) map[string]string {
	m := make(map[string]string, len(headers))
	for _, h := range headers {
		m[h.Key] = string(h.Value)
	}
	return m
}
