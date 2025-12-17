FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app ./internal/notification/cmd/main.go
FROM debian:bookworm-slim
WORKDIR /root/
COPY --from=builder /app/app .
COPY --from=builder /app/templates ./templates
CMD ["./app"]