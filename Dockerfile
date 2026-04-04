FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . /app

RUN apk add make git \
    && go install github.com/swaggo/swag/cmd/swag@latest \
    && make swagger \
    && make build

ENTRYPOINT ["/app/dermify-api", "version"]

FROM alpine

WORKDIR /app
COPY --from=builder /app/dermify-api /app
COPY --from=builder /app/config.yaml /app

ENTRYPOINT ["/app/dermify-api", "serve"]
