package internal

import (
	"net/http"
	"time"
	"webhookdelivery/internal/api"
	"webhookdelivery/internal/api/handlers"
	"webhookdelivery/internal/db/repositories"
	sqlc "webhookdelivery/internal/db/sqlc"
	"webhookdelivery/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	Addr string
	Db   *pgxpool.Pool
}

func NewApplication(addr string, db *pgxpool.Pool) *Application {
	return &Application{
		Addr: addr,
		Db:   db,
	}
}

func (app *Application) Run() error {
	router := api.NewRouter()

	queries := sqlc.New(app.Db)

	subcriptionRepository := repositories.NewSubscriptionRepository(queries)
	eventRepository := repositories.NewEventRepository(queries)
	deliveryRepository := repositories.NewDeliveryRepository(queries)

	eventService := services.NewEventService(app.Db, subcriptionRepository, eventRepository, deliveryRepository)

	router.RegisterHealthHandler(handlers.NewHealthHandler())
	router.RegisterSubscriptionHandler(handlers.NewSubscriptionHandler(subcriptionRepository))
	router.RegisterEventHandler(handlers.NewEventHandler(eventService))

	server := http.Server{
		Handler:      router,
		Addr:         app.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServe()
}
