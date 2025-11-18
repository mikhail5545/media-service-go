# --- Builder stage ---
# Use the official Go image as a builder.
# Using alpine for a smaller image size.
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files and setup dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application binary.
RUN CGO_ENABLED=0 GOOS=linux go build -o /media-service ./cmd/server/main.go

# --- Final stage ---
# Minimal base image for the final container
FROM alpine:latest

WORKDIR /

COPY --from=builder /media-service /media-service

CMD ["/media-service"]

# Expose necessary ports
EXPOSE 8083
EXPOSE 50053
