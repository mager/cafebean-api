package common

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
	"github.com/kelseyhightower/envconfig"
	"github.com/mager/cafebean-api/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Register registers all of the lifecycle methods and involkes the handler
func Register(
	lifecycle fx.Lifecycle,
	bq *bigquery.Client,
	database *firestore.Client,
	postgres *sql.DB,
	events *pubsub.Client,
	logger *zap.SugaredLogger,
	router *mux.Router,
) (
	*bigquery.Client,
	config.Config,
	*firestore.Client,
	*discordgo.Session,
	*pubsub.Client,
	*zap.SugaredLogger,
	*sql.DB,
	*mux.Router,
) {
	// Initialize config
	var cfg config.Config

	err := envconfig.Process("cafebean", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

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

	return bq, cfg, database, discord, events, logger, postgres, router
}
