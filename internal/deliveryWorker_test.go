package internal

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
	"webhookdelivery/internal/db/repositories"
	sqlc "webhookdelivery/internal/db/sqlc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestWorker(t *testing.T) (*DeliveryWorker, *repositories.MockDeliveryRepository) {
	t.Helper()

	deliveryRepository := repositories.NewMockDeliveryRepository(t)

	return NewDeliveryWorker(nil, deliveryRepository, nil), deliveryRepository
}

func TestProcessDeliveryResults_MarksSuccessfulDelivery(t *testing.T) {
	worker, deliveryRepository := newTestWorker(t)

	deliveryRepository.
		EXPECT().
		ProcessSuccessfulAttempt(mock.Anything, int64(1)).
		Return(nil).
		Once()

	worker.processDeliveryResults(context.Background(), []DeliveryResult{
		{delivery: sqlc.Delivery{ID: 1}, statusCode: http.StatusOK},
	})
}

func TestProcessDeliveryResults_ProcessesEveryResultAfterSuccess(t *testing.T) {
	worker, deliveryRepository := newTestWorker(t)

	deliveryRepository.
		EXPECT().
		ProcessSuccessfulAttempt(mock.Anything, int64(1)).
		Return(nil).
		Once()
	deliveryRepository.
		EXPECT().
		ProcessFailedAttempt(mock.Anything, int64(2), mock.Anything, int32(5), "unexpected status code 500").
		Return(nil).
		Once()
	deliveryRepository.
		EXPECT().
		ProcessSuccessfulAttempt(mock.Anything, int64(3)).
		Return(nil).
		Once()

	worker.processDeliveryResults(context.Background(), []DeliveryResult{
		{delivery: sqlc.Delivery{ID: 1}, statusCode: http.StatusOK},
		{delivery: sqlc.Delivery{ID: 2}, statusCode: http.StatusInternalServerError},
		{delivery: sqlc.Delivery{ID: 3}, statusCode: http.StatusNoContent},
	})
}

func TestProcessDeliveryResults_NonSuccessStatusDoesNotPanic(t *testing.T) {
	worker, deliveryRepository := newTestWorker(t)

	deliveryRepository.
		EXPECT().
		ProcessFailedAttempt(mock.Anything, int64(1), mock.Anything, int32(5), "unexpected status code 404").
		Return(nil).
		Once()

	require.NotPanics(t, func() {
		worker.processDeliveryResults(context.Background(), []DeliveryResult{
			{delivery: sqlc.Delivery{ID: 1}, statusCode: http.StatusNotFound},
		})
	})
}

func TestProcessDeliveryResults_NetworkErrorFailsDeliveryOnce(t *testing.T) {
	worker, deliveryRepository := newTestWorker(t)

	deliveryRepository.
		EXPECT().
		ProcessFailedAttempt(mock.Anything, int64(1), mock.Anything, int32(5), "connection refused").
		Return(nil).
		Once()

	worker.processDeliveryResults(context.Background(), []DeliveryResult{
		{delivery: sqlc.Delivery{ID: 1}, err: errors.New("connection refused")},
	})
}

func TestProcessDeliveryResults_NoResultsDoesNothing(t *testing.T) {
	worker, _ := newTestWorker(t)

	worker.processDeliveryResults(context.Background(), nil)
}

func TestCalculateNextAttempt_FirstAttemptDoesNotPanic(t *testing.T) {
	require.NotPanics(t, func() {
		calculateNextAttempt(0)
	})
}

func TestCalculateNextAttempt_FirstAttemptIsWithinBackoffWindow(t *testing.T) {
	delay := time.Until(calculateNextAttempt(0))

	assert.GreaterOrEqual(t, delay, 5*time.Second)
	assert.LessOrEqual(t, delay, 8*time.Second)
}

func TestCalculateNextAttempt_StaysInFutureForAnyAttemptCount(t *testing.T) {
	for _, attempts := range []int32{0, 1, 5, 31, 32, 40, 1000} {
		next := calculateNextAttempt(attempts)

		assert.Truef(
			t,
			next.After(time.Now()),
			"attempts=%d produced a delay of %v",
			attempts,
			time.Until(next),
		)
	}
}

func TestCalculateNextAttempt_BackoffGrowsAndIsCappedAtAnHour(t *testing.T) {
	first := time.Until(calculateNextAttempt(1))
	second := time.Until(calculateNextAttempt(2))

	assert.Greater(t, second, first)

	capped := time.Until(calculateNextAttempt(1000))

	assert.GreaterOrEqual(t, capped, time.Hour)
	assert.LessOrEqual(t, capped, 90*time.Minute)
}
