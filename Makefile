dev:
	go run main.go

test:
	go test ./...

build:
	gcloud builds submit --tag gcr.io/cafebean/cafebean-api

deploy:
	gcloud run deploy cafebean-api \
		--image gcr.io/cafebean/cafebean-api \
		--platform managed

ship:
	make test && make build && make deploy