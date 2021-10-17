package main

import (
	"database/sql"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
	bq "github.com/mager/cafebean-api/bigquery"
	"github.com/mager/cafebean-api/common"
	"github.com/mager/cafebean-api/database"
	"github.com/mager/cafebean-api/events"
	"github.com/mager/cafebean-api/handler"
	"github.com/mager/cafebean-api/logger"
	"github.com/mager/cafebean-api/postgres"
	"github.com/mager/cafebean-api/router"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			bq.Options,
			database.Options,
			postgres.Options,
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
	bq *bigquery.Client,
	database *firestore.Client,
	postgres *sql.DB,
	events *pubsub.Client,
	logger *zap.SugaredLogger,
	router *mux.Router,
) {
	bq, cfg, database, discord, events, logger, postgres, router := common.Register(
		lifecycle,
		bq,
		database,
		postgres,
		events,
		logger,
		router,
	)

	handler.New(
		bq,
		cfg,
		database,
		discord,
		events,
		logger,
		postgres,
		router,
	)
}
