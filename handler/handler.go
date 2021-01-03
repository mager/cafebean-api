package handler

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
	router *mux.Router
	logger *zap.SugaredLogger
	store  *firestore.Client
}

// New http handler
func New(router *mux.Router, logger *zap.SugaredLogger, store *firestore.Client) *Handler {
	h := Handler{router, logger, store}
	h.registerRoutes()

	return &h
}

// Bean represents a coffee bean
type Bean struct {
	Flavors []string `firestore:"flavors" json:"flavors" omitempty`
	Name    string   `firestore:"name" json:"name"`
	Roaster string   `firestore:"roaster" json:"roaster"`
	Shade   string   `firestore:"shade" json:"shade"`
}

// BeanDB represents a Bean in firestore
type BeanDB struct {
	Bean
}

// Roaster represents an organization that roasts beans
type Roaster struct {
	City     string         `firestore:"city" json:"city"`
	Location *latlng.LatLng `firestore:"location" json:"location"`
	Logo     string         `firestore:"logo" json:"logo"`
	Name     string         `firestore:"name" json:"name"`
	Slug     string         `firestore:"slug" json:"slug"`
	URL      string         `firestore:"url" json:"url"`
}

// RoasterDB represents a Roaster in firestore
type RoasterDB struct {
	Roaster
	Verified bool `firestore:"verified"`
}

// BeansResp is the response for the beans endpoint
type BeansResp struct {
	Beans []Bean `json:"beans"`
}

// RoastersResp is the response for the roasters endpoint
type RoastersResp struct {
	Roasters []Roaster `json:"roasters"`
}

// RegisterRoutes for all http endpoints
func (h *Handler) registerRoutes() {
	h.router.HandleFunc("/beans", h.getBeans).Methods("GET")
	h.router.HandleFunc("/beans", h.addBean).Methods("POST")
	h.router.HandleFunc("/roasters", h.getRoasters).Methods("GET")
}

func (h *Handler) getBeans(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &BeansResp{}
	)

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

type AddBeanReq struct {
	Flavors []string `json:"flavors"`
	Name    string   `json:"name"`
	Roaster string   `json:"roaster"`
	Shade   string   `json:"shade"`
}

type AddBeanResp struct {
	ID string `json:"id"`
}

func (h *Handler) addBean(w http.ResponseWriter, r *http.Request) {
	var (
		req  AddBeanReq
		resp = &AddBeanResp{}
		ctx  = context.TODO()
		err  error
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure roaster exists
	iter := h.store.Collection("roasters").Where("name", "==", req.Roaster).Documents(ctx)
	for {
		doc, err := iter.Next()
		if doc == nil {
			http.Error(w, "invalid roaster", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		break
	}

	// Add the bean
	doc, _, err := h.store.Collection("beans").Add(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.ID = doc.ID

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) getRoasters(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &RoastersResp{}
	)

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
