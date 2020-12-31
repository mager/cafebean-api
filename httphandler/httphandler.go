package httphandler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/type/latlng"
)

// Handler for http requests
type Handler struct {
	mux    *mux.Router
	logger *zap.SugaredLogger
	store  *firestore.Client
}

// New http handler
func New(s *mux.Router, logger *zap.SugaredLogger, store *firestore.Client) *Handler {
	h := Handler{s, logger, store}
	h.registerRoutes()

	return &h
}

type Bean struct {
	Name    string   `firestore:"name" json:"name"`
	Roaster string   `firestore:"roaster" json:"roaster"`
	Flavors []string `firestore:"flavors" json:"flavors"`
	Shade   string   `firestore:"shade" json:"shade"`
}
type Roaster struct {
	Name     string         `firestore:"name" json:"name"`
	Location *latlng.LatLng `firestore:"location" json:"location"`
	URL      string         `firestore:"url" json:"url"`
}

type BeansResp struct {
	Beans []Bean `json:"beans"`
}
type RoastersResp struct {
	Roasters []Roaster `json:"roasters"`
}

// RegisterRoutes for all http endpoints
func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/beans", h.beans)
	h.mux.HandleFunc("/roasters", h.roasters)
}

func (h *Handler) beans(w http.ResponseWriter, r *http.Request) {
	resp := &BeansResp{}

	// Call Firestore API
	iter := h.store.Collection("beans").Documents(context.TODO())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		var b Bean
		doc.DataTo(&b)
		resp.Beans = append(resp.Beans, b)
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) roasters(w http.ResponseWriter, r *http.Request) {
	resp := &RoastersResp{}

	// Call Firestore API
	iter := h.store.Collection("roasters").Documents(context.TODO())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		var r Roaster
		doc.DataTo(&r)
		resp.Roasters = append(resp.Roasters, r)
	}

	json.NewEncoder(w).Encode(resp)
}
