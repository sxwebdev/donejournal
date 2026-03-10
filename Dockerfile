# Admin frontend build stage
FROM node:25-alpine AS admin-frontend-builder

WORKDIR /app/frontend

RUN npm install -g pnpm

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ .
RUN pnpm run build

# Whisper.cpp — use official pre-built image (linux/amd64, linux/arm64)
# Includes whisper-cli, ffmpeg, and curl; no compilation needed
FROM ghcr.io/ggml-org/whisper.cpp:main AS whisper-source

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

# Build the application (CGO_ENABLED=0 for a fully static binary)
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-w -s -X 'main.version=${VERSION}' -X 'main.commitHash=${COMMIT_HASH}' -X 'main.buildDate=${BUILD_DATE}'" -o bin/donejournal ./cmd/donejournal

# Final stage — debian-slim for glibc compatibility with whisper-cli
FROM debian:bookworm-slim

# Install runtime dependencies: tzdata, ffmpeg for OGG→WAV conversion,
# libstdc++ and libgomp for whisper-cli shared library dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    tzdata \
    ffmpeg \
    libstdc++6 \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy whisper-cli and its whisper/ggml shared libraries from official image
COPY --from=whisper-source /app/build/bin/whisper-cli /usr/local/bin/whisper-cli
COPY --from=whisper-source /app/build/src/libwhisper.so.1 /usr/local/lib/libwhisper.so.1
COPY --from=whisper-source /app/build/ggml/src/libggml.so.0 /usr/local/lib/libggml.so.0
COPY --from=whisper-source /app/build/ggml/src/libggml-base.so.0 /usr/local/lib/libggml-base.so.0
COPY --from=whisper-source /app/build/ggml/src/libggml-cpu.so.0 /usr/local/lib/libggml-cpu.so.0
RUN ldconfig

# Copy Go binary
COPY --from=backend-builder /app/bin/donejournal .

ENTRYPOINT ["./donejournal"]
