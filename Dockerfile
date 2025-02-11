# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the service
RUN CGO_ENABLED=0 go build -o /connector-service ./go-server/cmd/server

# Final minimal image
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /connector-service /app/connector-service

EXPOSE 50051
ENTRYPOINT ["/app/connector-service"]
