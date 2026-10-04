package repositories

import (
	"context"

	sqlc "webhookdelivery/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
)

type EventRepository interface {
	WithTx(tx pgx.Tx) EventRepository
	CreateEvent(ctx context.Context, params sqlc.CreateEventParams) (sqlc.Event, error)
	GetEvent(ctx context.Context, id int64) (sqlc.Event, error)
}

type eventRepository struct {
	queries *sqlc.Queries
}

func NewEventRepository(queries *sqlc.Queries) EventRepository {
	return &eventRepository{
		queries: queries,
	}
}

func (r *eventRepository) WithTx(tx pgx.Tx) EventRepository {
	return &eventRepository{queries: r.queries.WithTx(tx)}
}

func (r *eventRepository) CreateEvent(ctx context.Context, params sqlc.CreateEventParams) (sqlc.Event, error) {
	event, err := r.queries.CreateEvent(ctx, params)
	if err != nil {
		return sqlc.Event{}, err
	}

	return event, nil
}

func (r *eventRepository) GetEvent(ctx context.Context, id int64) (sqlc.Event, error) {
	event, err := r.queries.GetEvent(ctx, id)
	if err != nil {
		return sqlc.Event{}, err
	}

	return event, nil
}
