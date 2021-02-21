package main

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
	"github.com/mager/cafebean-api/database"
	"github.com/mager/cafebean-api/events"
	"github.com/mager/cafebean-api/handler"
	"github.com/mager/cafebean-api/logger"
	"github.com/mager/cafebean-api/router"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			database.Options,
			events.Options,
			router.Options,
			logger.Options,
		),
		fx.Invoke(Register),
	).Run()
}

// Register registers all of the lifecycle methods and involkes the handler
func Register(
	lifecycle fx.Lifecycle,
	database *firestore.Client,
	events *pubsub.Client,
	logger *zap.SugaredLogger,
	router *mux.Router,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				logger.Info("Listening on :8080")
				go http.ListenAndServe(":8080", router)
				return nil
			},
			OnStop: func(context.Context) error {
				defer logger.Sync()
				defer database.Close()
				return nil
			},
		},
	)
	handler.New(database, events, logger, router)
}
