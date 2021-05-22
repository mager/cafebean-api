package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type AddProfileResp struct {
	User User `json:"user"`
}

// addProfile initializes the profile for the user.
// The initial payload comes from Auth0 and has a default nickname
// and profile photo.
// TODO: Add better security
func (h *Handler) addProfile(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		err       error
		req       UserDB
		resp      = &AddProfileResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "error handling request", http.StatusBadRequest)
		return
	}

	// Validate user
	if req.Email != userEmail {
		http.Error(w, "only the user can create their profile", http.StatusBadRequest)
		return
	}

	// Fetch the user first to make sure it doesn't exist
	iter := h.database.Collection("users").Where("email", "==", userEmail).Documents(ctx)
	for {
		doc, err := iter.Next()

		if doc != nil && err != nil {
			http.Error(w, "user already exists", http.StatusBadRequest)
			return
		}

		// Create a new user record if it doesn't exist
		newUser := UserDB{
			Email:     req.Email,
			Username:  req.Username,
			Photo:     req.Photo,
			CreatedAt: time.Now(),
		}
		created, _, err := h.database.Collection("users").Add(ctx, newUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.logger.Infow(
			"User added",
			"id", created.ID,
			"updated_by", userEmail,
		)

		resp.User = User{
			Photo:     newUser.Photo,
			Username:  newUser.Username,
			CreatedAt: newUser.CreatedAt,
		}

		break
	}

	json.NewEncoder(w).Encode(resp)
}
