run:
	export UNWEAVE_ENV="" && \
	export UNWEAVE_DOMAIN="" && \
	export UNWEAVE_API_URL"" UNWEAVE_APP_URL="" && \
	go run main.go

build:
	go build -o bin/unweave