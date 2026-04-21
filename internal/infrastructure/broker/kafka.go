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
	commandWriter kafka.Writer
	metricsWriter kafka.Writer
}

func NewBroker(addr string, cmdTopic, metricsTopic string) Broker {
	cmdWriter := kafka.Writer{
		Addr:         kafka.TCP(addr),
		Topic:        cmdTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}

	metWriter := kafka.Writer{
		Addr:         kafka.TCP(addr),
		Topic:        metricsTopic,
		Balancer:     &kafka.Hash{},
		Async:        true,
		BatchSize:    100,
		BatchTimeout: 500 * time.Millisecond,
	}

	return Broker{
		commandWriter: cmdWriter,
		metricsWriter: metWriter,
	}
}

func (b *Broker) SendStart(ctx context.Context, test domain.Test) error {
	cmd := testCommand{Action: "Start", Test: &test, TestID: test.ID}
	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal start command: %v", domain.ErrInternal, err)
	}

	if err := b.commandWriter.WriteMessages(ctx, kafka.Message{
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

	if err := b.commandWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(id.String()),
		Value: data,
	}); err != nil {
		return fmt.Errorf("%w: failed to write stop message: %v", domain.ErrInternal, err)
	}

	return nil
}

func (b *Broker) SendMetrics(ctx context.Context, metric domain.Metric) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal metric: %v", domain.ErrInternal, err)
	}

	if err := b.metricsWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(metric.TestID.String()),
		Value: data,
	}); err != nil {
		return fmt.Errorf("%w: failed to write metric message: %v", domain.ErrInternal, err)
	}

	return nil
}
