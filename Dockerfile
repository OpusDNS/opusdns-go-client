# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run tests
RUN go test -v ./...

# Build the library (optional - create example binary for validation)
RUN go build -o /dev/null ./...

# Final stage - minimal runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy test results or built artifacts if needed
COPY --from=builder /app /app

CMD ["echo", "OpusDNS Go client library built and tested successfully"]
