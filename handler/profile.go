package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

type User struct {
	Email    string `firestore:"email" json:"email"`
	Username string `firestore:"username" json:"username"`
	Location string `firestore:"location" json:"location"`
}

type UserReq struct {
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

	// Fetch the user
	iter := h.database.Collection("users").Where("email", "==", userEmail).Documents(ctx)
	for {
		doc, err := iter.Next()

		if doc == nil {
			// Create a new user record
			newUser := User{
				Email:    userEmail,
				Username: "",
				Location: "",
			}
			doc, _, err := h.database.Collection("users").Add(ctx, newUser)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			h.logger.Infow(
				"User added",
				"id", doc.ID,
				"updated_by", userEmail,
			)

			resp.User = newUser
		} else {
			var u User
			doc.DataTo(&u)
			resp.User = u
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		break
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		docID     string
		err       error
		req       UserReq
		resp      = &UserResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
			{Path: "email", Value: userEmail},
			{Path: "location", Value: req.Location},
			{Path: "username", Value: req.Username},
		},
	)
	h.logger.Infow(
		"User updated",
		"id", docsnap.Ref.ID,
		"updated_at", result.UpdateTime,
		"updated_by", userEmail,
	)

	resp.User = User{
		Username: req.Username,
		Location: req.Location,
		Email:    userEmail,
	}

	json.NewEncoder(w).Encode(resp)
}
