package bigquery

import (
	"context"
	"log"

	"cloud.google.com/go/bigquery"
)

// ProvideBQ provides a bigquery client
func ProvideBQ() *bigquery.Client {
	projectID := "cafebean"

	client, err := bigquery.NewClient(context.TODO(), projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return client
}

var Options = ProvideBQ
