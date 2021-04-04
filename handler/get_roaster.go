package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handler) getRoaster(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &RoasterResp{}
		vars = mux.Vars(r)
		slug = vars["slug"]
		ctx  = context.TODO()
	)

	// Get the roaster
	roasterIter := h.database.Collection("roasters").Where("slug", "==", slug).Documents(ctx)
	for {
		doc, err := roasterIter.Next()
		if doc == nil {
			http.Error(w, "invalid roaster", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.Roaster = docToRoaster(doc)

		break
	}

	// Get the beans for that roaster
	beansIter := h.database.Collection("beans").Where("roaster.slug", "==", resp.Roaster.Slug).Documents(ctx)
	for {
		doc, err := beansIter.Next()
		if doc == nil {
			break
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.Beans = append(resp.Beans, docToBean(doc))
	}

	json.NewEncoder(w).Encode(resp)
}
