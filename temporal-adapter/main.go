package main

import (
	"context"
	"log"

	"github.com/milovidov983/oms-temporal-demo/temporal-adapter/consumer"
	"github.com/milovidov983/oms-temporal-demo/temporal-adapter/handler"
)

func main() {
	cfg := consumer.ConsumerConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "oms-order-consumer",
		Topic:   "oms.oms-core.orders.v1",
		Handler: handler.NewOrderHandler(),
	}

	consumer, err := consumer.NewKafkaConsumer(cfg)
	if err != nil {
		log.Fatalf("Error creating consumer: %v", err)
	}

	ctx := context.Background()
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Error running consumer: %v", err)
	}
}
