package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

type Profile struct {
	Username string `json:"username"`
}

type ProfilePayload struct {
	User Profile `json:"user"`
}

type CreateProfileResp struct {
	User User `json:"user"`
}

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

// createProfile initializes the profile for the user.
// The initial payload comes from Auth0 and has a default nickname
// and profile photo.
// TODO: Add better security
func (h *Handler) createProfile(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		err       error
		req       UserDB
		resp      = &CreateProfileResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "error handling reques", http.StatusBadRequest)
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
			Email:    req.Email,
			Username: req.Username,
			Photo:    req.Photo,
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
			Photo:    newUser.Photo,
			Username: newUser.Username,
		}

		break
	}

	json.NewEncoder(w).Encode(resp)
}

// TODO: Add better security
func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
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

// checkUsername checks if a username is taken
func (h *Handler) checkUsername(w http.ResponseWriter, r *http.Request) {
	var (
		ctx      = context.TODO()
		vars     = mux.Vars(r)
		username = vars["username"]
	)

	q := h.database.Collection("users").Where("username", "==", username)
	iter := q.Documents(ctx)
	defer iter.Stop()
	for {
		_, err := iter.Next()

		// No user found, return 200
		if err == iterator.Done {
			return
		}

		// Error case
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// User found, return 400
		http.Error(w, "username taken", http.StatusBadRequest)
		return
	}
}
