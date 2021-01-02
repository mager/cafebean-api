package bundlefx

import (
	"context"
	"io"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"github.com/mager/caffy-beans/configfx"
	"github.com/mager/caffy-beans/firestorefx"
	"github.com/mager/caffy-beans/httpfx"
	"github.com/mager/caffy-beans/jaegerfx"
	"github.com/mager/caffy-beans/loggerfx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func registerHooks(
	lifecycle fx.Lifecycle,
	logger *zap.SugaredLogger,
	cfg *configfx.Config,
	store *firestore.Client,
	mux *mux.Router,
	tracingCloser io.Closer,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				logger.Info("Listening on ", cfg.ApplicationConfig.Address)
				go http.ListenAndServe(cfg.ApplicationConfig.Address, mux)
				return nil
			},
			OnStop: func(context.Context) error {
				defer store.Close()
				defer tracingCloser.Close()
				return logger.Sync()
			},
		},
	)
}

// Module provided to fx
var Module = fx.Options(
	configfx.Module,
	loggerfx.Module,
	firestorefx.Module,
	httpfx.Module,
	jaegerfx.Module,
	fx.Invoke(registerHooks),
)
