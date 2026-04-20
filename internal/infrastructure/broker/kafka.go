package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Guram-Gurych/LoadTest/internal/domain"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"time"
)

type testCommand struct {
	Action string       `json:"action"`
	Test   *domain.Test `json:"test,omitempty"`
	TestID uuid.UUID    `json:"test_id"`
}

type Broker struct {
	writer kafka.Writer
}

func NewBroker(addr string, topicName string, requiredAcks kafka.RequiredAcks, maxAttempts int, writeTimeout time.Duration) Broker {
	writer := kafka.Writer{
		Addr:         kafka.TCP(addr),
		Topic:        topicName,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: requiredAcks,
		MaxAttempts:  maxAttempts,
		WriteTimeout: writeTimeout,
	}

	return Broker{writer: writer}
}

func (b *Broker) SendStart(ctx context.Context, test domain.Test) error {
	cmd := testCommand{Action: "Start", Test: &test, TestID: test.ID}
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal start command: %v", domain.ErrInternal, err)
	}

	if err := b.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(test.ID.String()),
		Value: data,
	}); err != nil {
		return fmt.Errorf("%w: failed to write start message: %v", domain.ErrInternal, err)
	}

	return nil
}

func (b *Broker) SendStop(ctx context.Context, id uuid.UUID) error {
	cmd := testCommand{Action: "Stop", TestID: id}
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal stop command: %v", domain.ErrInternal, err)
	}

	if err := b.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(id.String()),
		Value: data,
	}); err != nil {
		return fmt.Errorf("%w: failed to write stop message: %v", domain.ErrInternal, err)
	}

	return nil
}
