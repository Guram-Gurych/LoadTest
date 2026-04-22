package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Guram-Gurych/LoadTest/internal/application"
	"github.com/Guram-Gurych/LoadTest/internal/config"
	kafkaDelivery "github.com/Guram-Gurych/LoadTest/internal/delivery/kafka"
	"github.com/Guram-Gurych/LoadTest/internal/infrastructure/broker"
)

func main() {
	cnf := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	kafkaBroker := broker.NewBroker(cnf.KafkaAddress, cnf.CommandTopic, cnf.MetricsTopic)

	workerSvc := application.NewWorkerService(ctx, &kafkaBroker)

	consumer := kafkaDelivery.NewConsumer(
		[]string{cnf.KafkaAddress},
		cnf.CommandTopic,
		"loadtest-worker-group",
		workerSvc,
		10,
		10e6,
	)

	log.Println("Worker is starting and listening to Kafka...")

	consumer.Run(ctx)

	log.Println("Worker stopped gracefully")
}
