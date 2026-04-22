package kafkaDelivery

import (
	"context"
	"encoding/json"
	"github.com/Guram-Gurych/LoadTest/internal/domain"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"log"
)

type testCommand struct {
	Action string       `json:"action"`
	Test   *domain.Test `json:"test,omitempty"`
	TestID uuid.UUID    `json:"test_id"`
}

type workerService interface {
	StartTest(ctx context.Context, test domain.Test)
	StopTest(ctx context.Context, id uuid.UUID) error
}

type Consumer struct {
	reader     *kafka.Reader
	service    workerService
	workersNum int
	commands   chan testCommand
}

func NewConsumer(brokers []string, topic string, groupID string, service workerService, workersNum, maxBytes int) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MaxBytes: maxBytes,
	})

	return &Consumer{
		reader:     reader,
		service:    service,
		workersNum: workersNum,
		commands:   make(chan testCommand),
	}
}

func (c *Consumer) Run(ctx context.Context) {
	for i := 0; i < c.workersNum; i++ {
		go c.workerLoop(ctx, i)
	}

	go c.readLoop(ctx)
}

func (c *Consumer) readLoop(ctx context.Context) {
	defer close(c.commands)

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			log.Printf("failed to read message: %v", err)
			continue
		}

		var cmd testCommand
		if err := json.Unmarshal(msg.Value, &cmd); err != nil {
			log.Printf("failed to unmarshal command: %v", err)
			continue
		}

		select {
		case <-ctx.Done():
			return
		case c.commands <- cmd:
		}
	}
}

func (c *Consumer) workerLoop(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd, ok := <-c.commands:
			if !ok {
				return
			}

			c.processCommand(ctx, cmd)
		}
	}
}

func (c *Consumer) processCommand(ctx context.Context, cmd testCommand) {
	switch cmd.Action {
	case "Start":
		if cmd.Test != nil {
			c.service.StartTest(ctx, *cmd.Test)
		}
	case "Stop":
		c.service.StopTest(ctx, cmd.TestID)
	default:
		log.Printf("Unknown action: %s", cmd.Action)
	}
}
