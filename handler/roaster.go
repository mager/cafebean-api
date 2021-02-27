package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"

	"google.golang.org/genproto/googleapis/type/latlng"
)

// Roaster represents an organization that roasts beans
type Roaster struct {
	City      string         `firestore:"city" json:"city"`
	Instagram string         `firestore:"instagram" json:"instagram"`
	Location  *latlng.LatLng `firestore:"location" json:"location"`
	Logo      string         `firestore:"logo" json:"logo"`
	Name      string         `firestore:"name" json:"name"`
	Slug      string         `firestore:"slug" json:"slug"`
	Twitter   string         `firestore:"twitter" json:"twitter"`
	URL       string         `firestore:"url" json:"url"`
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

// EditRoasterReq is the request body for editing a Roaster
type EditRoasterReq struct {
	City      string `firestore:"city" json:"city"`
	Instagram string `firestore:"instagram" json:"instagram"`
	Name      string `firestore:"name" json:"name"`
	Twitter   string `firestore:"twitter" json:"twitter"`
	URL       string `firestore:"url" json:"url"`
}

// EditRoasterResp is the response from the POST /roasters/{slug} endpoint
type EditRoasterResp struct {
	Roaster
}

func (h *Handler) editRoaster(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		vars      = mux.Vars(r)
		slug      = vars["slug"]
		err       error
		req       EditRoasterReq
		resp      = &EditRoasterResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the roaster
	roaster := h.database.Collection("roasters").Doc(slug)
	docsnap, err := roaster.Get(ctx)
	if err != nil {
		http.Error(w, "invalid roaster slug", http.StatusBadRequest)
		return
	}

	// Update the roaster
	result, err := roaster.Update(
		ctx,
		[]firestore.Update{
			{Path: "city", Value: req.City},
			{Path: "instagram", Value: req.Instagram},
			{Path: "name", Value: req.Name},
			{Path: "twitter", Value: req.Twitter},
			{Path: "url", Value: req.URL},
		},
	)
	h.logger.Infow(
		"Roaster updated",
		"id", docsnap.Ref.ID,
		"updated_at", result.UpdateTime,
		"updated_by", userEmail,
	)

	// // Send event
	// // TODO: Send updated fields
	t := h.events.Topic("roaster")
	res := t.Publish(ctx, &pubsub.Message{
		Data: []byte("Roaster updated"),
		Attributes: map[string]string{
			"id":         docsnap.Ref.ID,
			"user_email": userEmail,
		},
	})
	msgID, err := res.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	h.logger.Infow("Pubsub message succeeded", "msgId", msgID)

	// Send updated roaster response
	w.WriteHeader(http.StatusAccepted)

	updated, err := roaster.Get(ctx)
	if err != nil {
		h.logger.Errorw(
			"Error fetching roaster after updating it",
			"id", updated.Ref.ID,
		)
	}
	h.logger.Debug(updated)
	resp.Roaster = docToRoaster(updated)

	json.NewEncoder(w).Encode(resp)
}
