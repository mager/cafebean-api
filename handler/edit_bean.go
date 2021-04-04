package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

// EditBeanResp is the response from the POST /beans/{slug} endpoint
type EditBeanResp struct {
	Bean
}

func (h *Handler) editBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		docID     string
		vars      = mux.Vars(r)
		slug      = vars["slug"]
		err       error
		req       BeanReq
		resp      = &EditBeanResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the bean
	q := h.database.Collection("beans").Where("slug", "==", slug)
	iter := q.Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		docID = doc.Ref.ID
	}

	bean := h.database.Collection("beans").Doc(docID)
	docsnap, err := bean.Get(ctx)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, "invalid bean slug", http.StatusBadRequest)
		return
	}

	// Update the bean
	result, _ := bean.Update(
		ctx,
		[]firestore.Update{
			{Path: "countries", Value: req.Countries},
			{Path: "flavors", Value: req.Flavors},
			{Path: "description", Value: req.Description},
			{Path: "name", Value: req.Name},
			{Path: "photo", Value: req.Photo},
			{Path: "roaster.name", Value: req.Roaster.Name},
			{Path: "roaster.slug", Value: req.Roaster.Slug},
			{Path: "slug", Value: req.Slug},
			{Path: "url", Value: req.URL},
		},
	)
	h.logger.Infow(
		"Bean updated",
		"id", docsnap.Ref.ID,
		"updated_at", result.UpdateTime,
		"updated_by", userEmail,
	)

	// Publish an entry in BigQuery
	h.recordBeanChange(ctx, req, userEmail)

	// Send a webhook event to Discord
	msg, err := h.postBeanToDiscord(req, userEmail, "edit")
	h.logger.Info(msg)
	h.logger.Error(err)

	// Send updated bean response
	w.WriteHeader(http.StatusAccepted)

	updated, err := bean.Get(ctx)
	if err != nil {
		h.logger.Errorw(
			"Error fetching bean after updating it",
			"id", updated.Ref.ID,
		)
	}
	resp.Bean = docToBean(updated)

	json.NewEncoder(w).Encode(resp)
}
