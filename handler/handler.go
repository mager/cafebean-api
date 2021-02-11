package handler

import (
	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Handler for http requests
type Handler struct {
	logger   *zap.SugaredLogger
	router   *mux.Router
	database *firestore.Client
}

// ErrorMessage is a custom error message
type ErrorMessage struct {
	Message string `json:"message"`
}

// RegisterRoutes for all http endpoints
func (h *Handler) registerRoutes() {
	h.router.HandleFunc("/beans", h.getBeans).Methods("GET")
	h.router.HandleFunc("/beans", h.addBean).Methods("POST")
	h.router.HandleFunc("/beans/{slug}", h.getBean).Methods("GET")
	h.router.HandleFunc("/beans/{slug}", h.editBean).Methods("POST")
	h.router.HandleFunc("/roasters", h.getRoasters).Methods("GET")
}

// New http handler
func New(logger *zap.SugaredLogger, router *mux.Router, database *firestore.Client) *Handler {
	h := Handler{logger, router, database}
	h.registerRoutes()

	return &h
}
