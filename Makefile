dev:
	go run main.go

build:
	gcloud builds submit --tag gcr.io/cafebean/cafebean-api

deploy:
	gcloud run deploy --image gcr.io/cafebean/cafebean-api --platform managed --vpc-connector=cafebean-connector --vpc-egress=all

deploy-app-engine:
	gcloud app deploy