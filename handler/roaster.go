package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"

	"google.golang.org/genproto/googleapis/type/latlng"
)

// Roaster represents an organization that roasts beans
type Roaster struct {
	City     string         `firestore:"city" json:"city"`
	Location *latlng.LatLng `firestore:"location" json:"location"`
	Logo     string         `firestore:"logo" json:"logo"`
	Name     string         `firestore:"name" json:"name"`
	Slug     string         `firestore:"slug" json:"slug"`
	Twitter  string         `firestore:"twitter" json:"twitter"`
	URL      string         `firestore:"url" json:"url"`
}

// RoasterDB represents a Roaster in firestore
type RoasterDB struct {
	Roaster
	Verified bool `firestore:"verified"`
}

// RoasterResp is the response for the GET /roaster/{slug} endpoint
type RoasterResp struct {
	Roaster Roaster `json:"roaster"`
}

// RoastersResp is the response for the GET /roasters endpoint
type RoastersResp struct {
	Roasters []Roaster `json:"roasters"`
}

func docToRoaster(doc *firestore.DocumentSnapshot) Roaster {
	var r Roaster
	doc.DataTo(&r)
	r.Slug = doc.Ref.ID
	return r
}

func (h *Handler) getRoaster(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &RoasterResp{}
		vars = mux.Vars(r)
		slug = vars["slug"]
		ctx  = context.TODO()
	)

	// Get the bean
	doc, err := h.database.Collection("roasters").Doc(slug).Get(ctx)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(
			&ErrorMessage{
				Message: fmt.Sprintf("Failed to get document: %s", slug),
			},
		)
	} else {
		resp.Roaster = docToRoaster(doc)

		json.NewEncoder(w).Encode(resp)
	}
}

func (h *Handler) getRoasters(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &RoastersResp{}
	)

	// Call Firestore API
	iter := h.database.Collection("roasters").Documents(context.TODO())
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
