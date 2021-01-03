package db

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"go.uber.org/fx"
)

// ProvideDB provides a firestore client
func ProvideDB() *firestore.Client {
	projectID := "caffy-beans-api"

	client, err := firestore.NewClient(context.TODO(), projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client
}

// Module provided to fx
var Module = fx.Options(
	fx.Provide(ProvideDB),
)
