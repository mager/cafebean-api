package database

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

// ProvideDB provides a firestore client
func ProvideDB() *firestore.Client {
	projectID := "cafebean"

	client, err := firestore.NewClient(context.TODO(), projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client
}

var Options = ProvideDB
