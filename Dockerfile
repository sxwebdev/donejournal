# Backend build stage
FROM golang:1.25.3-alpine AS backend-builder

# Define build arguments for version, commit, and date.
ARG VERSION="unknown"
ARG COMMIT_HASH="unknown"
ARG BUILD_DATE="unknown"

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with embedded frontend
RUN go build -trimpath -ldflags="-w -s -X 'main.version=${VERSION}' -X 'main.commitHash=${COMMIT_HASH}' -X 'main.buildDate=${BUILD_DATE}'" -o bin/donejournal ./cmd/donejournal

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add tzdata

WORKDIR /app

# Copy binary from builder stage
COPY --from=backend-builder /app/bin/donejournal .

# Run the binary
ENTRYPOINT ["./donejournal"]
