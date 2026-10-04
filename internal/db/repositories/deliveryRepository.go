package repositories

import (
	"context"
	"time"
	sqlc "webhookdelivery/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type DeliveryRepository interface {
	WithTx(tx pgx.Tx) DeliveryRepository
	CreateDelivery(ctx context.Context, params sqlc.CreateDeliveryParams) (sqlc.Delivery, error)
	GetPendingDeliveries(ctx context.Context) ([]sqlc.Delivery, error)
	ChangeStatus(ctx context.Context, deliveryId int64, newStatus sqlc.DeliveryStatus) error
	ProcessFailedAttempt(
		ctx context.Context,
		deliveryId int64,
		nextAttemptAt time.Time,
		maxAttempts int32,
		lastError string) error
	ProcessSuccessfulAttempt(ctx context.Context, id int64) error
}

type deliveryRepository struct {
	queries *sqlc.Queries
}

func NewDeliveryRepository(queries *sqlc.Queries) DeliveryRepository {
	return &deliveryRepository{
		queries: queries,
	}
}

func (r *deliveryRepository) WithTx(tx pgx.Tx) DeliveryRepository {
	return &deliveryRepository{queries: r.queries.WithTx(tx)}
}

func (r *deliveryRepository) CreateDelivery(ctx context.Context, params sqlc.CreateDeliveryParams) (sqlc.Delivery, error) {
	delivery, err := r.queries.CreateDelivery(ctx, params)
	if err != nil {
		return sqlc.Delivery{}, err
	}

	return delivery, nil
}

func (r *deliveryRepository) GetPendingDeliveries(ctx context.Context) ([]sqlc.Delivery, error) {
	deliveries, err := r.queries.GetPendingDeliveries(ctx)
	if err != nil {
		return []sqlc.Delivery{}, err
	}

	return deliveries, nil
}

func (r *deliveryRepository) ChangeStatus(ctx context.Context, deliveryId int64, newStatus sqlc.DeliveryStatus) error {
	params := sqlc.ChangeStatusParams{
		ID:     deliveryId,
		Status: newStatus,
	}

	err := r.queries.ChangeStatus(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (r *deliveryRepository) ProcessFailedAttempt(
	ctx context.Context,
	deliveryId int64,
	nextAttemptAt time.Time,
	maxAttempts int32,
	lastError string) error {
	params := sqlc.ProcessFailedAttemptParams{
		ID: deliveryId,
		NextAttemptAt: pgtype.Timestamptz{
			Time:  nextAttemptAt,
			Valid: true,
		},
		LastError:   pgtype.Text{String: lastError, Valid: true},
		MaxAttempts: maxAttempts,
	}

	err := r.queries.ProcessFailedAttempt(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (r *deliveryRepository) ProcessSuccessfulAttempt(ctx context.Context, id int64) error {
	err := r.queries.ProcessSuccessfulAttempt(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
