package services

import (
	"context"
	"webhookdelivery/internal/db/repositories"
	sqlc "webhookdelivery/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventService interface {
	CreateEvent(ctx context.Context, params sqlc.CreateEventParams) (sqlc.Event, error)
}

type eventService struct {
	pool *pgxpool.Pool

	subscriptionRepository repositories.SubscriptionRepository
	eventRepository        repositories.EventRepository
	deliveryRepository     repositories.DeliveryRepository
}

func NewEventService(
	pool *pgxpool.Pool,
	subscriptionRepository repositories.SubscriptionRepository,
	eventRepository repositories.EventRepository,
	deliveryRepository repositories.DeliveryRepository) *eventService {
	return &eventService{
		pool:                   pool,
		subscriptionRepository: subscriptionRepository,
		eventRepository:        eventRepository,
		deliveryRepository:     deliveryRepository,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, params sqlc.CreateEventParams) (sqlc.Event, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return sqlc.Event{}, err
	}
	defer tx.Rollback(ctx)

	eventRepository := s.eventRepository.WithTx(tx)

	event, err := eventRepository.CreateEvent(ctx, params)
	if err != nil {
		return sqlc.Event{}, err
	}

	if err := s.createDeliveries(ctx, tx, &event); err != nil {
		return sqlc.Event{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Event{}, err
	}

	return event, nil
}

func (s *eventService) createDeliveries(ctx context.Context, tx pgx.Tx, event *sqlc.Event) error {
	subscriptionRepository := s.subscriptionRepository.WithTx(tx)
	deliveryRepository := s.deliveryRepository.WithTx(tx)

	urls, err := subscriptionRepository.GetUrls(ctx, event.EventType)
	if err != nil {
		return err
	}

	for _, url := range urls {
		_, err := deliveryRepository.CreateDelivery(
			ctx,
			sqlc.CreateDeliveryParams{
				EventID:     event.ID,
				EndpointUrl: url,
			})
		if err != nil {
			return err
		}
	}

	return nil
}
