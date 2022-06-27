.PHONY: bash dev start build

bash:
	@docker build -f Dockerfile-development -t shopline/mongodb-exporter . && docker run --rm --name mongodb_performance_exporter -it -v "$(shell pwd):/go/src" -p 3000:3000 --network bridge shopline/mongodb-exporter bash

dev:
	@docker build -f Dockerfile-development -t shopline/mongodb-exporter . && docker run --rm --name mongodb_performance_exporter -it -v "$(shell pwd):/go/src" -p 3000:3000  --network bridge shopline/mongodb-exporter air

start: build
	@./exporter

build:
	@go build -o exporter ./cmd/exporter.go
