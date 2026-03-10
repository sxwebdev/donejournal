# Admin frontend build stage
FROM node:25-alpine AS admin-frontend-builder

WORKDIR /app/frontend

RUN npm install -g pnpm

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ .
RUN pnpm run build

# Backend build stage
FROM golang:1.26.1-alpine AS backend-builder

# Define build arguments for version, commit, and date.
ARG VERSION="unknown"
ARG COMMIT_HASH="unknown"
ARG BUILD_DATE="unknown"

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy built frontend assets from the previous stage
COPY --from=admin-frontend-builder /app/frontend/dist ./frontend/dist

# Copy source code
COPY . .

# Build the application
RUN go build -trimpath -ldflags="-w -s -X 'main.version=${VERSION}' -X 'main.commitHash=${COMMIT_HASH}' -X 'main.buildDate=${BUILD_DATE}'" -o bin/donejournal ./cmd/donejournal

# Final stage — debian:slim for glibc compatibility with pre-built whisper binary
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    tzdata \
    ffmpeg \
    ca-certificates \
    wget \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Download pre-built whisper-cli binary for Linux x64 (glibc, no GPU)
# Uses official release from github.com/ggerganov/whisper.cpp
ARG WHISPER_VERSION=v1.8.3
RUN wget -q -O /tmp/whisper.zip \
    "https://github.com/ggerganov/whisper.cpp/releases/download/${WHISPER_VERSION}/whisper-bin-x64.zip" && \
    unzip -q /tmp/whisper.zip -d /tmp/whisper-bin && \
    install -m 755 /tmp/whisper-bin/whisper-cli /usr/local/bin/whisper-cli && \
    rm -rf /tmp/whisper.zip /tmp/whisper-bin

WORKDIR /app

# Copy binary from backend builder stage
COPY --from=backend-builder /app/bin/donejournal .

# Run the binary
ENTRYPOINT ["./donejournal"]
