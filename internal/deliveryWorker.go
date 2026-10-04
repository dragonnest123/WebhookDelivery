package internal

import (
	"bytes"
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
	"webhookdelivery/internal/db/repositories"
	"webhookdelivery/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DeliveryWorker struct {
	db                 *pgxpool.Pool
	deliveryRepository repositories.DeliveryRepository
	eventRepository    repositories.EventRepository

	client       http.Client
	workInterval time.Duration
	maxAttempts  int32
}

type DeliveryResult struct {
	delivery   sqlc.Delivery
	statusCode int
	err        error
}

func NewDeliveryWorker(
	pool *pgxpool.Pool,
	deliveryRepository repositories.DeliveryRepository,
	eventRepository repositories.EventRepository) *DeliveryWorker {
	return &DeliveryWorker{
		db:                 pool,
		deliveryRepository: deliveryRepository,
		eventRepository:    eventRepository,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		workInterval: time.Second,
		maxAttempts:  5,
	}
}

func (worker *DeliveryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(worker.workInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			worker.work(ctx)
		}
	}
}

func (worker *DeliveryWorker) work(ctx context.Context) {
	tx, err := worker.db.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	deliveryRepository := worker.deliveryRepository.WithTx(tx)
	eventRepository := worker.eventRepository.WithTx(tx)

	deliveries, err := deliveryRepository.GetPendingDeliveries(ctx)
	if err != nil {
		return
	}

	urls := make([]string, 0, len(deliveries))
	payloads := make([][]byte, 0, len(deliveries))
	for _, delivery := range deliveries {
		if err := deliveryRepository.ChangeStatus(ctx, delivery.ID, sqlc.DeliveryStatusProcessing); err != nil {
			return
		}

		event, err := eventRepository.GetEvent(ctx, delivery.EventID)
		if err != nil {
			return
		}

		urls = append(urls, delivery.EndpointUrl)
		payloads = append(payloads, event.Payload)
	}

	if err := tx.Commit(ctx); err != nil {
		return
	}

	result := worker.sendToSubscribers(deliveries, urls, payloads)
	worker.processDeliveryResults(ctx, result)
}

func (worker *DeliveryWorker) sendToSubscribers(deliveries []sqlc.Delivery, urls []string, payloads [][]byte) []DeliveryResult {
	result := []DeliveryResult{}

	for i := range urls {
		response, err := worker.client.Post(urls[i], "application/json", bytes.NewReader(payloads[i]))
		if err != nil {
			result = append(result, DeliveryResult{
				delivery: deliveries[i],
				err:      err})
			continue
		}
		response.Body.Close()

		result = append(result, DeliveryResult{
			delivery:   deliveries[i],
			statusCode: response.StatusCode,
			err:        nil})
	}

	return result
}

func (worker *DeliveryWorker) processDeliveryResults(ctx context.Context, results []DeliveryResult) {
	for _, result := range results {
		if result.err != nil {
			worker.processFailedDelivery(ctx, result, result.err.Error())
			continue
		}

		if result.statusCode >= 200 && result.statusCode < 300 {
			worker.deliveryRepository.ProcessSuccessfulAttempt(ctx, result.delivery.ID)
			continue
		}

		worker.processFailedDelivery(ctx, result, fmt.Sprintf("unexpected status code %d", result.statusCode))
	}
}

func (worker *DeliveryWorker) processFailedDelivery(ctx context.Context, result DeliveryResult, lastError string) {
	worker.deliveryRepository.ProcessFailedAttempt(
		ctx,
		result.delivery.ID,
		calculateNextAttempt(result.delivery.Attempts),
		worker.maxAttempts,
		lastError,
	)
}

func calculateNextAttempt(attempts int32) time.Time {
	exponent := min(max(attempts, 1), 16)

	delay := min(5*time.Second*time.Duration(1<<(exponent-1)), time.Hour)

	jitterMin := delay * 20 / 100
	jitterMax := delay * 50 / 100

	delay += time.Duration(rand.Int64N(int64(jitterMax-jitterMin) + 1))

	return time.Now().Add(delay)
}
