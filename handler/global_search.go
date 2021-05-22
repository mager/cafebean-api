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
	Only  string `json:"only"`
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
	Name    string   `json:"name"`
	Slug    string   `json:"slug"`
	Flavors []string `json:"flavors"`
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
		// of a roaster slug (replacing dashes with spaces)
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

	// Fetch all the beans
	if req.Only != "roaster" {
		beansIter := h.database.Collection("beans").Documents(ctx)
		for {
			doc, err := beansIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				h.logger.Fatalf("Failed to iterate: %v", err)
			}

			b := docToBean(doc)

			// Check if the search query is a substring of a slug or a "spaced" version
			// of a bean slug (replacing dashes with spaces)
			spacedSlug := strings.Replace(b.Slug, "-", " ", -1)
			if strings.Contains(b.Slug, query) || strings.Contains(spacedSlug, query) {
				resp.Results = append(resp.Results, GlobalSearchResult{
					Roaster: GlobalSearchRoaster{
						Name: b.Roaster.Name,
						Slug: b.Roaster.Slug,
					},
					Bean: &GlobalSearchBean{
						Name:    b.Name,
						Slug:    b.Slug,
						Flavors: b.Flavors,
					},
				})
			}

			// Check if the search query matches any of the bean flavors
			for _, f := range b.Flavors {
				if f == query {
					resp.Results = append(resp.Results, GlobalSearchResult{
						Roaster: GlobalSearchRoaster{
							Name: b.Roaster.Name,
							Slug: b.Roaster.Slug,
						},
						Bean: &GlobalSearchBean{
							Name:    b.Name,
							Slug:    b.Slug,
							Flavors: b.Flavors,
						},
					})
				}
			}
		}
	}

	json.NewEncoder(w).Encode(resp)
}
