package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

// TODO: Add better security
func (h *Handler) editProfile(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		docID     string
		err       error
		req       ProfilePayload
		resp      = &ProfilePayload{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := req.User.Username

	// Fetch the user
	iter := h.database.Collection("users").Where("email", "==", userEmail).Documents(ctx)
	for {
		doc, err := iter.Next()

		if doc == nil {
			http.Error(w, "invalid user", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate user
		var u UserDB
		doc.DataTo(&u)
		if u.Email != userEmail {
			http.Error(w, "only the user can update their profile", http.StatusBadRequest)
			return
		}

		docID = doc.Ref.ID

		break
	}

	// Update the user
	user := h.database.Collection("users").Doc(docID)
	docsnap, err := user.Get(ctx)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, "invalid user", http.StatusBadRequest)
		return
	}

	result, _ := user.Update(
		ctx,
		[]firestore.Update{
			{Path: "username", Value: username},
		},
	)
	h.logger.Infow(
		"User updated",
		"id", docsnap.Ref.ID,
		"username", username,
		"updated_at", result.UpdateTime,
		"updated_by", userEmail,
	)

	resp.User.Username = username

	json.NewEncoder(w).Encode(resp)
}
