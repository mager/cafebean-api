package handler

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

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
