package handler

import (
	"encoding/json"
	"net/http"
)

// GetReviewsResp is the response for the GET /reviews endpoint
type GetReviewsResp struct {
	Reviews []ReviewWithBean `json:"reviews"`
}

func (h *Handler) getReviews(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &GetReviewsResp{}
		// ctx      = context.TODO()
		// beanDocs = []*firestore.DocumentRef{}
		// reviews  []ReviewWithBean
		// beans    = h.database.Collection("beans")
	)

	// Call Postgres
	// rows, err := h.postgres.Query(`
	// 	SELECT
	// 		r.review_id,
	// 		r.rating,
	// 		r.review,
	// 		r.bean_ref,
	// 		r.updated_at,
	// 		u.username as user
	// 	FROM reviews r
	// 	LEFT JOIN users u on r.user_id = u.user_id;
	// `)
	// if err != nil {
	// 	h.logger.Error(err)
	// }
	// defer rows.Close()
	// for rows.Next() {
	// 	var reviewId int
	// 	var rating float64
	// 	var review string
	// 	var beanRef string
	// 	var updatedAt time.Time
	// 	var user string
	// 	err = rows.Scan(&reviewId, &rating, &review, &beanRef, &updatedAt, &user)
	// 	if err != nil {
	// 		h.logger.Error(err)
	// 	}
	// 	h.logger.Infow(
	// 		"Result from Postgres",
	// 		"reviewId", reviewId,
	// 		"beanRef", beanRef,
	// 		"review", review,
	// 		"rating", rating,
	// 		"updatedAt", updatedAt,
	// 		"user", user,
	// 	)

	// 	reviews = append(reviews, ReviewWithBean{
	// 		Review:    review,
	// 		Rating:    rating,
	// 		UpdatedAt: updatedAt,
	// 		User:      user,
	// 	})
	// 	beanDocs = append(beanDocs, beans.Doc(beanRef))
	// }
	// err = rows.Err()
	// if err != nil {
	// 	h.logger.Error(err)
	// }

	// // Fetch related beans
	// beanSnaps, err := h.database.GetAll(ctx, beanDocs)
	// if err != nil {
	// 	// TODO: Handle error.
	// 	h.logger.Error(err)
	// }

	// for i, review := range reviews {
	// 	review.Bean = docToBean(beanSnaps[i])
	// 	resp.Reviews = append(resp.Reviews, review)
	// }

	json.NewEncoder(w).Encode(resp)
}
