package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
)

// GetBeanResp is the response for the GET /bean/{slug} endpoint
type GetBeanResp struct {
	Bean    Bean     `json:"bean"`
	Reviews []Review `json:"reviews"`
}

func (h *Handler) getBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx     = context.TODO()
		resp    = &GetBeanResp{}
		vars    = mux.Vars(r)
		slug    = vars["slug"]
		beanDoc *firestore.DocumentSnapshot
		err     error
		reviews []Review
	)
	// Get the bean
	beanIter := h.database.Collection("beans").Where("slug", "==", slug).Documents(ctx)
	for {
		beanDoc, err = beanIter.Next()
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

	h.logger.Info(beanDoc.Ref.ID)

	if h.cfg.ReviewsEnabled {
		// Get reviews
		rows, err := h.postgres.Query(`
		SELECT
			r.review_id,
			r.rating,
			r.review,
			r.bean_ref,
			r.updated_at,
			u.username as user
		FROM reviews r
		LEFT JOIN users u on r.user_id = u.user_id
		WHERE r.bean_ref = $1;
	`, beanDoc.Ref.ID)
		if err != nil {
			h.logger.Error(err)
		}
		defer rows.Close()
		for rows.Next() {
			var reviewId int
			var rating float64
			var review string
			var beanRef string
			var updatedAt time.Time
			var user string
			err = rows.Scan(&reviewId, &rating, &review, &beanRef, &updatedAt, &user)
			if err != nil {
				h.logger.Error(err)
			}
			h.logger.Infow(
				"Result from Postgres",
				"reviewId", reviewId,
				"beanRef", beanRef,
				"review", review,
				"rating", rating,
				"updatedAt", updatedAt,
				"user", user,
			)

			reviews = append(reviews, Review{
				Review:    review,
				Rating:    rating,
				UpdatedAt: updatedAt,
				User:      user,
				Bean:      slug,
			})
		}
		err = rows.Err()
		if err != nil {
			h.logger.Error(err)
		}

		resp.Reviews = reviews
	}

	json.NewEncoder(w).Encode(resp)
}
