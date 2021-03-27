package handler

import (
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"github.com/mager/cafebean-api/config"
	"go.uber.org/zap"
)

// Handler for http requests
type Handler struct {
	bq       *bigquery.Client
	cfg      config.Config
	database *firestore.Client
	discord  *discordgo.Session
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
	// Stats
	h.router.HandleFunc("/stats", h.getStats).Methods("GET")

	// Beans
	h.router.HandleFunc("/beans", h.getBeans).Methods("GET")
	h.router.HandleFunc("/beans", h.addBean).Methods("POST")
	h.router.HandleFunc("/beans/{slug}", h.getBean).Methods("GET")
	h.router.HandleFunc("/beans/{slug}", h.editBean).Methods("POST")

	// Roasters
	h.router.HandleFunc("/roasters", h.getRoasters).Methods("GET")
	h.router.HandleFunc("/roasters", h.addRoaster).Methods("POST")
	h.router.HandleFunc("/roasters/{slug}", h.getRoaster).Methods("GET")
	h.router.HandleFunc("/roasters/{slug}", h.editRoaster).Methods("POST")
	h.router.HandleFunc("/roasters_list", h.getRoastersList).Methods("GET")

	// Users
	h.router.HandleFunc("/profile", h.getProfile).Methods("GET")
	h.router.HandleFunc("/profile", h.updateProfile).Methods("POST")
}

// New http handler
func New(
	bq *bigquery.Client,
	cfg config.Config,
	database *firestore.Client,
	discord *discordgo.Session,
	events *pubsub.Client,
	logger *zap.SugaredLogger,
	router *mux.Router,
) *Handler {
	h := Handler{bq, cfg, database, discord, events, logger, router}
	h.registerRoutes()

	return &h
}
