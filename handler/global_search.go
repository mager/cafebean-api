package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/api/iterator"
)

type GlobalSearchReq struct {
	Query string `json:"query"`
}

type GlobalSearchResp struct {
	Results []GlobalSearchResult `json:"results"`
}

type GlobalSearchResult struct {
	Bean    *GlobalSearchBean   `json:"bean,omitempty"`
	Roaster GlobalSearchRoaster `json:"roaster"`
}

type GlobalSearchRoaster struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}
type GlobalSearchBean struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// globalSearch initializes the profile for the user.
func (h *Handler) globalSearch(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		err       error
		req       GlobalSearchReq
		userEmail = r.Header.Get("X-User-Email")
		resp      = &GlobalSearchResp{}
		query     string
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "error handling request", http.StatusBadRequest)
		return
	}

	query = strings.ToLower(req.Query)

	h.logger.Infof("New global search request from %s for %s", userEmail, query)

	// Fetch all the roasters
	roasterIter := h.database.Collection("roasters").Documents(ctx)
	for {
		doc, err := roasterIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			h.logger.Fatalf("Failed to iterate: %v", err)
		}

		r := docToRoaster(doc)

		// Check if the search query is a substring of a slug or a "spaced" version
		// of the slug (replacing dashes with spaces)
		spacedSlug := strings.Replace(r.Slug, "-", " ", -1)
		if strings.Contains(r.Slug, query) || strings.Contains(spacedSlug, query) {
			resp.Results = append(resp.Results, GlobalSearchResult{
				Roaster: GlobalSearchRoaster{
					Name: r.Name,
					Slug: r.Slug,
				},
				Bean: nil,
			})
		}

	}
	json.NewEncoder(w).Encode(resp)
}
