package events

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
)

// ProvideEvents provides a pubsub client
func ProvideEvents() *pubsub.Client {
	projectID := "cafebean"

	client, err := pubsub.NewClient(context.TODO(), projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client
}

var Options = ProvideEvents
