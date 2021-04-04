package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handler) getBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = context.TODO()
		resp = &BeanResp{}
		vars = mux.Vars(r)
		slug = vars["slug"]
	)

	// Get the bean
	iter := h.database.Collection("beans").Where("slug", "==", slug).Documents(ctx)
	for {
		doc, err := iter.Next()
		if doc == nil {
			http.Error(w, "invalid bean", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.Bean = docToBean(doc)

		break
	}

	json.NewEncoder(w).Encode(resp)
}
