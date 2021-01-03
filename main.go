package main

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"github.com/mager/caffy-beans/config"
	"github.com/mager/caffy-beans/db"
	"github.com/mager/caffy-beans/handler"
	"github.com/mager/caffy-beans/logger"
	"github.com/mager/caffy-beans/router"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Options(
			config.Module,
			db.Module,
			router.Module,
			logger.Module,
		),
		fx.Invoke(Register),
	).Run()
}

func Register(
	lifecycle fx.Lifecycle,
	cfg *config.Config,
	logger *zap.SugaredLogger,
	router *mux.Router,
	store *firestore.Client,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				logger.Info("Listening on ", cfg.ApplicationConfig.Address)
				go http.ListenAndServe(cfg.ApplicationConfig.Address, router)
				return nil
			},
			OnStop: func(context.Context) error {
				defer store.Close()
				defer logger.Sync()
				return nil
			},
		},
	)
	handler.New(router, logger, store)
}
