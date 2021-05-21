package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// Stats represents stats
type Stats struct {
	BeanCount    int            `firestore:"bean_count" json:"bean_count"`
	RoasterCount int            `firestore:"roaster_count" json:"roaster_count"`
	FlavorMap    map[string]int `json:"flavor_map"`
}

type StatsResp struct {
	Stats Stats `json:"stats"`
}

func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	var (
		flavorMap = make(map[string]int)
		resp      = &StatsResp{}
		ctx       = context.TODO()
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

	// Get flavor map
	for _, bean := range beans {
		for _, flavor := range docToBean(bean).Flavors {
			f := strings.ToLower(flavor)
			_, ok := flavorMap[f]
			if ok {
				flavorMap[f] += 1
			} else {
				flavorMap[f] = 1
			}
		}
	}
	resp.Stats.FlavorMap = flavorMap

	json.NewEncoder(w).Encode(resp)
}
