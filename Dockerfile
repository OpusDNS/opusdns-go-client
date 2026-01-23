# Build stage
FROM golang:1.25-alpine AS builder

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

# Final stage - Go runtime for tests
FROM golang:1.25-alpine

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy source code for running tests
COPY --from=builder /app /app

# Re-download dependencies (go mod cache not copied)
RUN go mod download

CMD ["go", "test", "-v", "./..."]
