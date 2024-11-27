package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/milovidov983/oms-temporal-demo/oms-core/repository"
	"github.com/milovidov983/oms-temporal-demo/shared/events"
	"github.com/milovidov983/oms-temporal-demo/shared/models"
)

type AssemblyApplicationService struct {
	repo  repository.AssemblyApplicationRepository
	kafka sarama.SyncProducer
	topic string
}

func NewAssemblyApplicationService(
	repo repository.AssemblyApplicationRepository,
	kafka sarama.SyncProducer,
	topic string,
) *AssemblyApplicationService {
	return &AssemblyApplicationService{
		repo:  repo,
		kafka: kafka,
		topic: topic,
	}
}

func (s *AssemblyApplicationService) CreateAssemblyApplication(
	ctx context.Context,
	orderID string,
) (*models.AssemblyApplication, error) {
	application, err := s.repo.Create(ctx, orderID)

	if err != nil {
		return nil, fmt.Errorf("failed to save assembly application: %w", err)
	}
	log.Printf("[debug] assembly application with ID %s saved to database", application.ID)

	if err := s.publishAssemblyApplication(application); err != nil {
		return nil, fmt.Errorf("failed to publish assembly application event: %w", err)
	}

	return application, nil
}

func (s *AssemblyApplicationService) CompleteAssembly(
	ctx context.Context,
	applicationID string,
) error {
	application, err := s.repo.Complete(ctx, applicationID)
	if err != nil {
		return fmt.Errorf("failed to complete assembly: %w", err)
	}
	log.Printf("[debug] assembly application with ID %s completed", applicationID)

	if err := s.publishAssemblyApplicationCompleted(application); err != nil {
		return fmt.Errorf("failed to publish assembly application completed event: %w", err)
	}

	return nil
}

func (s *AssemblyApplicationService) CancelAssembly(
	ctx context.Context,
	applicationID string,
) error {
	err := s.repo.Cancel(ctx, applicationID)
	if err != nil {
		return fmt.Errorf("failed to get assembly application: %w", err)
	}
	log.Printf("[debug] assembly application with ID %s canceled", applicationID)

	if err := s.publishAssemblyApplicationCanceled(applicationID); err != nil {
		return fmt.Errorf("failed to publish assembly application canceled event: %w", err)
	}

	return nil
}

func (s *AssemblyApplicationService) publishAssemblyApplicationCanceled(
	applicationID string,
) error {
	log.Printf("[debug] publishing assembly application canceled event for ID %s", applicationID)
	return nil
}

func (s *AssemblyApplicationService) publishAssemblyApplication(
	application *models.AssemblyApplication,
) error {
	log.Printf("[debug] publishing assembly application event for ID %s and status: %v and order %s", application.ID, application.Status, application.OrderID)
	return nil
}

func (s *AssemblyApplicationService) publishAssemblyApplicationCompleted(
	application *models.AssemblyApplication,
) error {
	log.Printf("[debug] publishing assembly application event for ID %s and status: %v and order %s", application.ID, application.Status, application.OrderID)

	event := &events.AssemblyApplicationEvent{
		EventType: events.AssemblyCompleted,
		EventData: events.AssemblyEventData{
			ID:      application.ID,
			OrderID: application.OrderID,
		},
	}

	msg, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal assembly event: %w", err)
	}

	partition, offset, err := s.kafka.SendMessage(&sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.StringEncoder(msg),
	})

	log.Printf("[debug] event published to kafka topic %s, partition %d, offset %d", s.topic, partition, offset)

	return err
}
