package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/api/iterator"
)

func (h *Handler) getBeans(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &BeansResp{}
		ctx  = context.TODO()
	)

	// Call Firestore API
	iter := h.database.Collection("beans").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			h.logger.Fatalf("Failed to iterate: %v", err)
		}

		resp.Beans = append(resp.Beans, docToBean(doc))
	}

	json.NewEncoder(w).Encode(resp)
}
