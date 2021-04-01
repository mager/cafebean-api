package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	Photo    string `json:"photo"`
	Username string `json:"username"`
}

type UserDB struct {
	Email    string `firestore:"email" json:"email"`
	Photo    string `firestore:"photo" json:"photo"`
	Username string `firestore:"username" json:"username"`
}

type PrivateUser struct {
	User UserDB `json:"user"`
}

type UserResp struct {
	User User `json:"user"`
}

// getUser fetches public user information
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		vars      = mux.Vars(r)
		username  = vars["username"]
		resp      = &UserResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	h.logger.Infow(
		"User viewed",
		"username", username,
		"requested_by", userEmail,
	)

	// Fetch the user
	iter := h.database.Collection("users").Where("username", "==", username).Documents(ctx)
	for {
		doc, err := iter.Next()

		if doc == nil {
			http.Error(w, "user not found", http.StatusNotFound)
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
