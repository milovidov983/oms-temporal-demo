package consumer

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/milovidov983/oms-temporal-demo/shared/events"
)

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	topics   Topics
	logger   *log.Logger
	handler  EventHandler
}

type EventHandler interface {
	HandleOrderEvent(event events.OrderEvent)
	HandleAssemblyApplicationEvent(event events.AssemblyApplicationEvent)
}

type ConsumerConfig struct {
	Brokers []string
	GroupID string
	Topics  Topics
	Handler EventHandler
}

type Topics struct {
	Orders               string
	AssemblyApplications string
}

func (t *Topics) ToStringArray() []string {
	var result []string

	val := reflect.ValueOf(t).Elem()

	for i := 0; i < val.NumField(); i++ {
		result = append(result, val.Field(i).String())
	}

	return result
}

func (cfg *ConsumerConfig) Check() {
	if len(cfg.Brokers) == 0 {
		log.Fatal("[fatal] Kafka Consumer Brokers are not set")
	}
	if cfg.GroupID == "" {
		log.Fatal("[fatal] Kafka Consumer GroupID is not set")
	}
	for _, topic := range cfg.Topics.ToStringArray() {
		if topic == "" {
			log.Fatal("[fatal] Kafka Consumer Topic is not set")
		}
	}
	if cfg.Handler == nil {
		log.Fatal("[fatal] Kafka Consumer Handler is not set")
	}
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
		topics:   cfg.Topics,
		logger:   log.New(os.Stdout, "kafka-consumer: ", log.LstdFlags),
		handler:  cfg.Handler,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	topics := c.topics.ToStringArray()
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := c.consumer.Consume(ctx, topics, c); err != nil {
				c.logger.Printf("[error] Error from consumer: %v", err)
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
		c.logger.Println("[info] Terminating: context cancelled")
	case <-sigterm:
		c.logger.Println("[info] Terminating: via signal")
	}

	cancel := func() {}
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg.Wait()
	if err := c.consumer.Close(); err != nil {
		c.logger.Printf("[error] Error closing consumer: %v", err)
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

func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		switch message.Topic {
		case c.topics.Orders:
			c.logger.Printf("[info] Received message from topic %s", message.Topic)
			c.handleOrderTopic(message)
		case c.topics.AssemblyApplications:
			c.logger.Printf("[info] Received message from topic %s", message.Topic)
			c.handleAssemblyApplicationsTopic(message)
		default:
			c.logger.Printf("[info] Received message from unknown topic %s", message.Topic)
		}

		session.MarkMessage(message, "")
	}
	return nil
}

func (c *KafkaConsumer) handleOrderTopic(message *sarama.ConsumerMessage) {
	var event events.OrderEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		c.logger.Printf("[error] handleOrderTopic error unmarshaling message: %v", err)
		return
	}

	c.handler.HandleOrderEvent(event)
}

func (c *KafkaConsumer) handleAssemblyApplicationsTopic(message *sarama.ConsumerMessage) {
	var event events.AssemblyApplicationEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		c.logger.Printf("[error] handleAssemblyApplicationsTopic error unmarshaling message: %v", err)
		return
	}

	c.handler.HandleAssemblyApplicationEvent(event)
}
