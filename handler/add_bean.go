package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// AddBeanReq is the request body for adding a Bean
type AddBeanReq struct {
	Bean
}

// AddBeanResp is the response from the POST /beans endpoint
type AddBeanResp struct {
	ID string `json:"id"`
}

func (h *Handler) addBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		err       error
		req       BeanReq
		resp      = &AddBeanResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure roaster exists
	iter := h.database.Collection("roasters").Where("name", "==", req.Roaster.Name).Documents(ctx)
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
	doc, _, err := h.database.Collection("beans").Add(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Infow(
		"Bean added",
		"id", doc.ID,
		"updated_by", userEmail,
	)

	resp.ID = doc.ID

	// Publish an entry in BigQuery
	h.recordBeanChange(ctx, req, userEmail)

	// Send a webhook event to Discord
	msg, err := h.postBeanToDiscord(req, userEmail, "add")
	h.logger.Info(msg)
	h.logger.Error(err)

	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(resp)
}
