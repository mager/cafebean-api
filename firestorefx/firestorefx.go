package firestorefx

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"go.uber.org/fx"
)

func ProvideFirestore() *firestore.Client {
	projectID := "caffy-api"

	client, err := firestore.NewClient(context.TODO(), projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client
}

// Module provided to fx
var Module = fx.Options(
	fx.Provide(ProvideFirestore),
)
