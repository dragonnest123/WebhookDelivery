package services_test

import (
	"context"
	"errors"
	"testing"
	"webhookdelivery/internal"
	"webhookdelivery/internal/db/repositories"
	sqlc "webhookdelivery/internal/db/sqlc"
	"webhookdelivery/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testEventType   = "order.created"
	testEndpointUrl = "https://example.com/hook"
)

func countRows(t *testing.T, db *pgxpool.Pool, query string) int {
	t.Helper()

	var count int
	require.NoError(t, db.QueryRow(context.Background(), query).Scan(&count))

	return count
}

func prepareSubscription(t *testing.T, db *pgxpool.Pool, queries *sqlc.Queries) {
	t.Helper()

	remove := func() {
		_, _ = db.Exec(
			context.Background(),
			"DELETE FROM subscriptions WHERE event_type = $1 AND endpoint_url = $2",
			testEventType,
			testEndpointUrl,
		)
	}

	remove()
	t.Cleanup(remove)

	_, err := queries.CreateSubscription(context.Background(), sqlc.CreateSubscriptionParams{
		EventType:   testEventType,
		EndpointUrl: testEndpointUrl,
	})
	require.NoError(t, err)
}

func TestCreateEvent_ShouldSaveEvent(t *testing.T) {
	db := internal.ConnectToTestDb()
	t.Cleanup(db.Close)

	queries := sqlc.New(db)

	service := services.NewEventService(
		db,
		repositories.NewSubscriptionRepository(queries),
		repositories.NewEventRepository(queries),
		repositories.NewDeliveryRepository(queries),
	)

	event, err := service.CreateEvent(
		context.Background(),
		sqlc.CreateEventParams{
			EventType: testEventType,
			Payload:   []byte(`{"order_id": 123}`),
		},
	)

	require.NoError(t, err)
	require.NotZero(t, event.ID)

	savedEvent, err := queries.GetEvent(context.Background(), event.ID)

	require.NoError(t, err)
	require.Equal(t, event.ID, savedEvent.ID)
	require.Equal(t, testEventType, savedEvent.EventType)
}

func TestCreateEvent_RollbackWhenDeliveryCreationFails(t *testing.T) {
	db := internal.ConnectToTestDb()
	t.Cleanup(db.Close)

	queries := sqlc.New(db)

	prepareSubscription(t, db, queries)

	deliveryRepository := repositories.NewMockDeliveryRepository(t)

	deliveryRepository.
		EXPECT().
		WithTx(mock.Anything).
		Return(deliveryRepository)

	deliveryRepository.
		EXPECT().
		CreateDelivery(mock.Anything, mock.MatchedBy(func(params sqlc.CreateDeliveryParams) bool {
			return params.EndpointUrl == testEndpointUrl
		})).
		Return(sqlc.Delivery{}, errors.New("failed to create delivery"))

	service := services.NewEventService(
		db,
		repositories.NewSubscriptionRepository(queries),
		repositories.NewEventRepository(queries),
		deliveryRepository,
	)

	eventsBefore := countRows(t, db, "SELECT count(*) FROM events")
	deliveriesBefore := countRows(t, db, "SELECT count(*) FROM deliveries")

	event, err := service.CreateEvent(
		context.Background(),
		sqlc.CreateEventParams{
			EventType: testEventType,
			Payload:   []byte(`{"order_id": 123}`),
		},
	)

	require.Error(t, err)
	require.Zero(t, event.ID)

	require.Equal(t, eventsBefore, countRows(t, db, "SELECT count(*) FROM events"))
	require.Equal(t, deliveriesBefore, countRows(t, db, "SELECT count(*) FROM deliveries"))
}
