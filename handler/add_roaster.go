package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// AddRoasterResp is the response from the POST /roasters/{slug} endpoint
type AddRoasterResp struct {
	ID string `json:"id"`
}

func (h *Handler) addRoaster(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		err       error
		req       RoasterReq
		resp      = &AddRoasterResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure the roaster doesn't already exist
	roasterIter := h.database.Collection("roasters").Where("slug", "==", req.Slug).Documents(ctx)
	for {
		doc, err := roasterIter.Next()

		if doc != nil {
			http.Error(w, "roaster already exists", http.StatusBadRequest)
			return
		}
		if err != nil && err.Error() != "no more items in iterator" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		break
	}

	// Add the roaster
	doc, _, err := h.database.Collection("roasters").Add(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infow(
		"Roaster added",
		"id", doc.ID,
		"updated_by", userEmail,
	)

	// Publish an entry in BigQuery
	h.recordRoasterChange(ctx, req, userEmail)

	// Send a webhook event to Discord
	h.postRoasterToDiscord(req, userEmail, "add")

	// Send updated roaster response
	w.WriteHeader(http.StatusAccepted)

	resp.ID = doc.ID

	json.NewEncoder(w).Encode(resp)
}
