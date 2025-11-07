FROM golang:1.25 AS builder
  WORKDIR /build
  COPY . .
  RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd/media_managment_service

FROM alpine:latest
  WORKDIR /app
  COPY --from=builder /build/service .
  COPY --from=builder /build/cmd/media_managment_service/conf/*.yaml ./config/
  ENV CONFIG_PATH=/app/config
  EXPOSE 8080
  CMD ["./service"]
