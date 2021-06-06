package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// Stats represents stats
type Stats struct {
	BeanCount        int               `firestore:"bean_count" json:"bean_count"`
	RoasterCount     int               `firestore:"roaster_count" json:"roaster_count"`
	RoasterLocations []RoasterLocation `json:"roaster_locations"`
}

type RoasterLocation struct {
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
	Name string  `json:"name"`
	Slug string  `json:"slug"`
}

type StatsResp struct {
	Stats Stats `json:"stats"`
}

func (h *Handler) getStats(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &StatsResp{}
		ctx  = context.TODO()
	)

	// Get bean count
	beans, err := h.database.Collection("beans").Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp.Stats.BeanCount = len(beans)

	// Get roaster count
	roasters, err := h.database.Collection("roasters").Documents(ctx).GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp.Stats.RoasterCount = len(roasters)

	// Get roaster locations
	for _, roaster := range roasters {
		r := docToRoaster(roaster)
		resp.Stats.RoasterLocations = append(resp.Stats.RoasterLocations, RoasterLocation{
			Lat:  r.Location.Latitude,
			Lng:  r.Location.Longitude,
			Name: r.Name,
			Slug: r.Slug,
		})
	}

	json.NewEncoder(w).Encode(resp)
}
