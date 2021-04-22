package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/api/iterator"
)

// GetReviewsResp is the response for the GET /reviews endpoint
type GetReviewsResp struct {
	Reviews []ReviewWithBean `json:"reviews"`
}

func (h *Handler) getReviews(w http.ResponseWriter, r *http.Request) {
	var (
		resp      = &GetReviewsResp{}
		ctx       = context.TODO()
		beanSlugs []string
		reviews   []ReviewWithBean
		beans     []Bean
	)

	// Get all reviews
	reviewsIter := h.database.Collection("reviews").Documents(ctx)
	for {
		doc, err := reviewsIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			h.logger.Fatalf("Failed to iterate: %v", err)
		}

		data := doc.Data()
		beanSlugs = append(beanSlugs, data["bean"].(string))
		reviews = append(reviews, docToReviewWithBean(doc))
	}

	// Fetch related beans
	beansIter := h.database.Collection("beans").Where("slug", "in", beanSlugs).Documents(ctx)
	for {
		doc, err := beansIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			h.logger.Fatalf("Failed to iterate: %v", err)
		}

		beans = append(beans, docToBean(doc))
	}

	// Populate response
	for i, review := range reviews {
		review.Bean = beans[i]
		resp.Reviews = append(resp.Reviews, review)
	}

	json.NewEncoder(w).Encode(resp)
}
