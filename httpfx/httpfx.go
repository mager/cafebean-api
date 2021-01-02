package httpfx

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/fx"
)

func ProvideHTTP() *mux.Router {
	var router = mux.NewRouter()
	router.Use(jsonMiddleware)
	return router
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Module provided to fx
var Module = fx.Options(
	fx.Provide(ProvideHTTP),
)
