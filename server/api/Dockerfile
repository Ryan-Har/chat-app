# Define the Go version
FROM golang:1.22.1-alpine AS builder

# Set working directory
WORKDIR /go/src/app

# Copy application code and dependencies
COPY . .

# Install dependencies using Go modules
RUN go mod download

# Build the application
RUN go build -o main .

# Define the final image
FROM alpine:latest AS User

RUN mkdir /app

# Copy the built binary
COPY --from=builder /go/src/app/main /app

# Set working directory
WORKDIR /app

# Define the entrypoint
ENTRYPOINT ["./main"]

EXPOSE 8001