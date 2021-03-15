package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// Stats represents stats
type Stats struct {
	BeanCount    int `firestore:"bean_count" json:"bean_count"`
	RoasterCount int `firestore:"roaster_count" json:"roaster_count"`
}

type StatsResp struct {
	Stats Stats `json:"stats"`
}

func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &StatsResp{}
		ctx  = context.TODO()
	)

	beans, err := h.database.Collection("beans").Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp.Stats.BeanCount = len(beans)

	roasters, err := h.database.Collection("roasters").Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp.Stats.RoasterCount = len(roasters)

	json.NewEncoder(w).Encode(resp)
}
