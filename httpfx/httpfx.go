package httpfx

import (
	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

// Module provided to fx
var Module = fx.Options(
	fx.Provide(mux.NewRouter),
	// fx.Provide(http.NewServeMux),
)
