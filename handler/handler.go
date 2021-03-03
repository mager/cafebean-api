package handler

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Handler for http requests
type Handler struct {
	database *firestore.Client
	events   *pubsub.Client
	logger   *zap.SugaredLogger
	router   *mux.Router
}

// ErrorMessage is a custom error message
type ErrorMessage struct {
	Message string `json:"message"`
}

// RegisterRoutes for all http endpoints
func (h *Handler) registerRoutes() {
	// Beans
	h.router.HandleFunc("/beans", h.getBeans).Methods("GET")
	h.router.HandleFunc("/beans", h.addBean).Methods("POST")
	h.router.HandleFunc("/beans/{slug}", h.getBean).Methods("GET")
	h.router.HandleFunc("/beans/{slug}", h.editBean).Methods("POST")

	// Roasters
	h.router.HandleFunc("/roasters", h.getRoasters).Methods("GET")
	h.router.HandleFunc("/roasters/{slug}", h.getRoaster).Methods("GET")
	h.router.HandleFunc("/roasters/{slug}", h.editRoaster).Methods("POST")
	h.router.HandleFunc("/roasters_list", h.getRoastersList).Methods("GET")
}

// New http handler
func New(
	database *firestore.Client,
	events *pubsub.Client,
	logger *zap.SugaredLogger,
	router *mux.Router,
) *Handler {
	h := Handler{database, events, logger, router}
	h.registerRoutes()

	return &h
}
