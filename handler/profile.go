package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type User struct {
	Email    string `firestore:"email" json:"email"`
	Username string `firestore:"username" json:"username"`
	Location string `firestore:"location" json:"location"`
}

type UserResp struct {
	User User `json:"user"`
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		resp      = &UserResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	resp.User.Email = userEmail
	h.logger.Info(userEmail)
	// Fetch the user from Firestore
	iter := h.database.Collection("users").Where("email", "==", userEmail).Documents(ctx)
	for {
		doc, err := iter.Next()
		if doc == nil {
			http.Error(w, "user does not exist", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var u User
		doc.DataTo(&u)
		resp.User = u
		break
	}

	json.NewEncoder(w).Encode(resp)
}
