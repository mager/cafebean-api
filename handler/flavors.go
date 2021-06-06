package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
)

// FlavorsResp represents bean flavors
type FlavorsResp struct {
	Flavors map[string]int `json:"flavors"`
}

func (h *Handler) getFlavorMap(beans []*firestore.DocumentSnapshot) map[string]int {
	var flavorMap = make(map[string]int)

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

	return flavorMap
}

func (h *Handler) getFlavors(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &FlavorsResp{}
		ctx  = context.TODO()
	)

	// Get bean count
	beans, err := h.database.Collection("beans").Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get flavor map
	resp.Flavors = h.getFlavorMap(beans)

	json.NewEncoder(w).Encode(resp)
}
