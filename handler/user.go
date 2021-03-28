package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type UserDB struct {
	Email    string `firestore:"email" json:"email"`
	Location string `firestore:"location" json:"location"`
	Username string `firestore:"username" json:"username"`
}

type PublicUser struct {
	Location string `json:"location"`
	Username string `json:"username"`
}

type PrivateUser struct {
	User UserDB `json:"user"`
}

type UserResp struct {
	User PublicUser `json:"user"`
}

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
		var u PublicUser
		doc.DataTo(&u)
		resp.User = u

		break
	}
	json.NewEncoder(w).Encode(resp)
}
