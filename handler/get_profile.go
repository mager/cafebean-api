package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

type GetProfileResp struct {
	User UserDB `json:"user"`
}

// getProfile fetches the user's private profile info
// TODO: Add better security
func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		resp      = &GetProfileResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	resp.User.Email = userEmail

	// Fetch the user
	iter := h.database.Collection("users").Where("email", "==", userEmail).Documents(ctx)
	for {
		doc, err := iter.Next()

		if doc == nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		var u UserDB
		doc.DataTo(&u)
		resp.User = u

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		break
	}
	json.NewEncoder(w).Encode(resp)
}
