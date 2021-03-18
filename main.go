package main

import (
	"context"
	"net/http"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
	beelineClient "github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/wrappers/hnynethttp"
	"github.com/mager/cafebean-api/beeline"
	bq "github.com/mager/cafebean-api/bigquery"
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
			beeline.Options,
			bq.Options,
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
	beelineConfig beelineClient.Config,
	bq *bigquery.Client,
	database *firestore.Client,
	events *pubsub.Client,
	logger *zap.SugaredLogger,
	router *mux.Router,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				logger.Info("Listening on :8080")
				beelineClient.Init(beelineConfig)
				go http.ListenAndServe(":8080", hnynethttp.WrapHandler(router))
				return nil
			},
			OnStop: func(context.Context) error {
				defer logger.Sync()
				defer database.Close()
				defer beelineClient.Close()
				return nil
			},
		},
	)
	handler.New(bq, database, events, logger, router)
}
