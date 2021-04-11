package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
)

// GetBeanResp is the response for the GET /bean/{slug} endpoint
type GetBeanResp struct {
	BeanPath string       `json:"bean_path"`
	Bean     Bean         `json:"bean"`
	Reviews  []BeanReview `json:"reviews"`
}

func (h *Handler) getBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx        = context.TODO()
		resp       = &GetBeanResp{}
		vars       = mux.Vars(r)
		slug       = vars["slug"]
		beanDocRef *firestore.DocumentRef
	)

	// Get the bean
	beanIter := h.database.Collection("beans").Where("slug", "==", slug).Documents(ctx)
	for {
		beanDoc, err := beanIter.Next()
		if beanDoc == nil {
			http.Error(w, "invalid bean", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.Bean = docToBean(beanDoc)
		beanDocRef = beanDoc.Ref
		break
	}

	// Get reviews
	reviewsIter := h.database.Collection("reviews").Where("bean", "==", beanDocRef).Documents(ctx)
	for {
		reviewsDoc, err := reviewsIter.Next()
		if reviewsDoc == nil {
			break
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var r BeanReview
		reviewsDoc.DataTo(&r)
		resp.Reviews = append(resp.Reviews, r)

		break
	}

	json.NewEncoder(w).Encode(resp)
}
