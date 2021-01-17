package main

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"github.com/mager/caffy-beans/config"
	"github.com/mager/caffy-beans/database"
	"github.com/mager/caffy-beans/handler"
	"github.com/mager/caffy-beans/logger"
	"github.com/mager/caffy-beans/router"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			config.Options,
			database.Options,
			router.Options,
			logger.Options,
		),
		fx.Invoke(Register),
	).Run()
}

func Register(
	lifecycle fx.Lifecycle,
	database *firestore.Client,
	cfg *config.Config,
	logger *zap.SugaredLogger,
	router *mux.Router,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				logger.Info("Listening on ", cfg.Application.Address)
				go http.ListenAndServe(cfg.Application.Address, router)
				return nil
			},
			OnStop: func(context.Context) error {
				defer logger.Sync()
				defer database.Close()
				return nil
			},
		},
	)
	handler.New(logger, router, database)
}
