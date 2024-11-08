package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/milovidov983/oms-temporal-demo/oms-core/repository"
	"github.com/milovidov983/oms-temporal-demo/shared/events"
	"github.com/milovidov983/oms-temporal-demo/shared/models"
)

type OrderService struct {
	repo  *repository.OrderRepository
	kafka sarama.SyncProducer
	topic string
}

func NewOrderService(repo *repository.OrderRepository, kafka sarama.SyncProducer, topic string) *OrderService {
	return &OrderService{
		repo:  repo,
		kafka: kafka,
		topic: topic,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *models.Order) error {
	if err := s.validateOrder(order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	order.Status = models.OrderStatusCreated
	order.ID = uuid.New().String()

	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	log.Printf("[debug] order with ID %s saved to database", order.ID)

	if err := s.publishOrderCreated(order); err != nil {
		return fmt.Errorf("failed to publish order created event: %w", err)
	}

	return nil
}

func (s *OrderService) GetOrderStatus(ctx context.Context, orderID string) (models.OrderStatus, error) {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return "", fmt.Errorf("failed to get order: %w", err)
	}

	return order.Status, nil
}

func (s *OrderService) validateOrder(order *models.Order) error {
	return nil
}

func (s *OrderService) publishOrderCreated(order *models.Order) error {
	event := &events.OrderEvent{
		EventType: events.OrderCreated,
		EventData: *order,
	}
	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}

	partition, offset, err := s.kafka.SendMessage(&sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.StringEncoder(msg),
	})

	log.Printf("[debug] event published to kafka topic %s, partition %d, offset %d", s.topic, partition, offset)

	return err
}
