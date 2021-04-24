package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	bq "github.com/mager/cafebean-api/bigquery"
	"github.com/mager/cafebean-api/config"
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
			config.Options,
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
	cfg config.Config,
	database *firestore.Client,
	postgres *sql.DB,
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

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", cfg.DiscordAuthToken))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	handler.New(bq, cfg, database, discord, events, logger, postgres, router)
}
