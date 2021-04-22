package handler

import (
	"time"

	"cloud.google.com/go/firestore"
)

// Review is a review of a bean by a user
type Review struct {
	Rating    float64   `firestore:"rating" json:"rating"`
	Review    string    `firestore:"review" json:"review"`
	User      string    `firestore:"user" json:"user"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
	Bean      string    `firestore:"bean" json:"bean"`
}

type ReviewWithBean struct {
	Rating    float64   `firestore:"rating" json:"rating"`
	Review    string    `firestore:"review" json:"review"`
	User      string    `firestore:"user" json:"user"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
	Bean      Bean      `firestore:"bean" json:"bean"`
}

func docToReview(doc *firestore.DocumentSnapshot) Review {
	var r Review
	doc.DataTo(&r)
	return r
}

func docToReviewWithBean(doc *firestore.DocumentSnapshot) ReviewWithBean {
	var r ReviewWithBean
	doc.DataTo(&r)
	return r
}
