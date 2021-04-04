package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

// EditRoasterResp is the response from the POST /roasters/{slug} endpoint
type EditRoasterResp struct {
	Roaster
}

func (h *Handler) editRoaster(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		docID     string
		vars      = mux.Vars(r)
		slug      = vars["slug"]
		err       error
		req       RoasterReq
		resp      = &EditRoasterResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the roaster
	q := h.database.Collection("roasters").Where("slug", "==", slug)
	roasterIter := q.Documents(ctx)
	defer roasterIter.Stop()
	for {
		doc, err := roasterIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		docID = doc.Ref.ID
	}

	roaster := h.database.Collection("roasters").Doc(docID)
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
			{Path: "location", Value: req.Location},
			{Path: "logo", Value: req.Logo},
			{Path: "name", Value: req.Name},
			{Path: "slug", Value: req.Slug},
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

	// Publish an entry in BigQuery
	h.recordRoasterChange(ctx, req, userEmail)

	// Send a webhook event to Discord
	h.postRoasterToDiscord(req, userEmail, "edit")

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
