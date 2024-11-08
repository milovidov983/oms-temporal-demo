package consumer

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/milovidov983/oms-temporal-demo/shared/events"
)

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	logger   *log.Logger
	handler  EventHandler
}

type EventHandler interface {
	HandleOrderCreated(event events.OrderEvent) error
	HandleOrderCancelled(event events.OrderEvent) error
}

type ConsumerConfig struct {
	Brokers []string
	GroupID string
	Topic   string
	Handler EventHandler
}

func NewKafkaConsumer(cfg ConsumerConfig) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategySticky()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: consumer,
		topic:    cfg.Topic,
		logger:   log.New(os.Stdout, "kafka-consumer: ", log.LstdFlags),
		handler:  cfg.Handler,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	topics := []string{c.topic}
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := c.consumer.Consume(ctx, topics, c); err != nil {
				c.logger.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	c.logger.Println("Kafka consumer started")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		c.logger.Println("Terminating: context cancelled")
	case <-sigterm:
		c.logger.Println("Terminating: via signal")
	}

	cancel := func() {}
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg.Wait()
	if err := c.consumer.Close(); err != nil {
		c.logger.Printf("Error closing consumer: %v", err)
		return err
	}

	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var event events.OrderEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Printf("Error unmarshaling message: %v", err)
			continue
		}

		var err error
		switch event.EventType {
		case events.OrderCreated:
			err = c.handler.HandleOrderCreated(event)
		case events.OrderCancelled:
			err = c.handler.HandleOrderCancelled(event)
		default:
			c.logger.Printf("Unknown event type: %s", event.EventType)
			continue
		}

		if err != nil {
			c.logger.Printf("Error handling event: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}
