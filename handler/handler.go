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

// New http handler
func New(logger *zap.SugaredLogger, router *mux.Router, database *firestore.Client) *Handler {
	h := Handler{logger, router, database}
	h.registerRoutes()

	return &h
}
