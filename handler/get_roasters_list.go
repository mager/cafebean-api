package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"google.golang.org/api/iterator"
)

func (h *Handler) getRoastersList(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &RoastersListResp{}
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

		var r RoasterMap
		doc.DataTo(&r)

		resp.Roasters = append(resp.Roasters, r)
	}

	json.NewEncoder(w).Encode(resp)
}
