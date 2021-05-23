package handler

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
	bq "github.com/mager/cafebean-api/bigquery"
	"github.com/mager/cafebean-api/common"
	"github.com/mager/cafebean-api/database"
	"github.com/mager/cafebean-api/events"
	"github.com/mager/cafebean-api/logger"
	"github.com/mager/cafebean-api/postgres"
	"github.com/mager/cafebean-api/router"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// Register registers all of the lifecycle methods and involkes the handler
// It's copied from main.go
func Register(
	lifecycle fx.Lifecycle,
	bq *bigquery.Client,
	database *firestore.Client,
	postgres *sql.DB,
	events *pubsub.Client,
	logger *zap.SugaredLogger,
	router *mux.Router,
) {
	bq, cfg, database, discord, events, logger, postgres, router := common.Register(
		lifecycle,
		bq,
		database,
		postgres,
		events,
		logger,
		router,
	)

	New(bq, cfg, database, discord, events, logger, postgres, router)
}

func Test_globalSearch(t *testing.T) {
	type test struct {
		name  string
		query string
		only  string
		exp   string
	}

	tests := []test{
		{
			name:  "roaster only matches slug",
			query: "ipsento",
			only:  "roaster",
			exp:   "{\"results\":[{\"roaster\":{\"name\":\"Ipsento\",\"slug\":\"ipsento\"}}]}\n",
		},
		{
			name:  "beans matches slug",
			query: "cascade",
			exp:   "{\"results\":[{\"bean\":{\"name\":\"Cascade Espresso\",\"slug\":\"ipsento-cascade-espresso\",\"flavors\":[\"dark chocolate\",\"mixed nuts\"]},\"roaster\":{\"name\":\"Ipsento\",\"slug\":\"ipsento\"}}]}\n",
		},
		{
			name:  "flavors match",
			query: "jordan almond",
			exp:   "{\"results\":[{\"bean\":{\"name\":\"Jumpstart\",\"slug\":\"partners-coffee-jumpstart\",\"flavors\":[\"caramel\",\"jordan almond\",\"poached pear\"]},\"roaster\":{\"name\":\"Partners Coffee\",\"slug\":\"partners-coffee\"}}]}\n",
		},
	}

	testApp := fxtest.New(t,
		fx.Provide(
			bq.Options,
			database.Options,
			postgres.Options,
			events.Options,
			router.Options,
			logger.Options,
		),
		fx.Invoke(Register),
	)
	defer testApp.RequireStart().RequireStop()

	for _, tc := range tests {
		// perform setUp before each test here
		t.Run(tc.name, func(t *testing.T) {
			t.Log(tc.query)
			var jsonStr = []byte(fmt.Sprintf(`{"query":"%s","only":"%s"}`, tc.query, tc.only))

			req, _ := http.NewRequest("POST", "http://localhost:8080/search", bytes.NewBuffer(jsonStr))
			req.Header.Set("X-User-Email", "test@cafebean.org")
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("Received non-200 response: %d\n", resp.StatusCode)
			}
			body, _ := ioutil.ReadAll(resp.Body)

			assert.Equal(t, tc.exp, string(body))
		})
	}
}
