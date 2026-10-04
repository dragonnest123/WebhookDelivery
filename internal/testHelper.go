package internal

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToTestDb() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/webhookdelivery_test")
	if err != nil {
		panic(err)
	}

	return pool
}
