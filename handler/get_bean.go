package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

// GetBeanResp is the response for the GET /bean/{slug} endpoint
type GetBeanResp struct {
	Bean    Bean     `json:"bean"`
	Reviews []Review `json:"reviews"`
}

func (h *Handler) getBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = context.TODO()
		resp = &GetBeanResp{}
		vars = mux.Vars(r)
		slug = vars["slug"]
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
		break
	}

	// Get reviews
	q := h.database.Collection("reviews").Where("bean", "==", slug)
	iter := q.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data := doc.Data()
		resp.Reviews = append(resp.Reviews, Review{
			Rating:    data["rating"].(float64),
			Review:    data["review"].(string),
			UpdatedAt: data["updated_at"].(time.Time),
			User:      data["user"].(string),
		})
	}

	json.NewEncoder(w).Encode(resp)
}
