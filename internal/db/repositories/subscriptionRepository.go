package repositories

import (
	"context"
	sqlc "webhookdelivery/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
)

type SubscriptionRepository interface {
	WithTx(tx pgx.Tx) SubscriptionRepository
	CreateSubscription(ctx context.Context, params sqlc.CreateSubscriptionParams) (sqlc.Subscription, error)
	GetUrls(ctx context.Context, eventType string) ([]string, error)
}

type subscriptionRepository struct {
	queries *sqlc.Queries
}

func NewSubscriptionRepository(queries *sqlc.Queries) SubscriptionRepository {
	return &subscriptionRepository{
		queries: queries,
	}
}

func (r *subscriptionRepository) WithTx(tx pgx.Tx) SubscriptionRepository {
	return &subscriptionRepository{queries: r.queries.WithTx(tx)}
}

func (r *subscriptionRepository) CreateSubscription(ctx context.Context, params sqlc.CreateSubscriptionParams) (sqlc.Subscription, error) {
	subscription, err := r.queries.CreateSubscription(ctx, params)
	if err != nil {
		return sqlc.Subscription{}, err
	}

	return subscription, nil
}

func (r *subscriptionRepository) GetUrls(ctx context.Context, eventType string) ([]string, error) {
	urls, err := r.queries.GetUrlsByEventType(ctx, eventType)
	if err != nil {
		return []string{}, err
	}

	return urls, nil
}
