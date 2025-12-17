FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app ./internal/hotel-srv/cmd/main.go
FROM debian:bookworm-slim
WORKDIR /root/
COPY --from=builder /app/app .
CMD ["./app"]