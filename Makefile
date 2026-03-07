BINARY_NAME = dermify-api
IMAGE_NAME = dermify-api

build:
	@go build -o $(BINARY_NAME) -ldflags="-X main.Commit=$(git rev-parse HEAD)" .

build-image:
	@docker build -t $(IMAGE_NAME) .

run-image:
	@docker run -p 8080:8080 $(IMAGE_NAME)

lint:
	@golangci-lint run --config golangci.yaml

lint-fix:
	@golangci-lint run --config golangci.yaml --fix

test:
	@go test ./... -count=1 -v

test-short:
	@go test ./internal/... -count=1 -short

swagger:
	@swag init -g internal/api/api.go --parseDependency --parseInternal -o docs
