# Stage 1: Build the binary
FROM golang:1.24.5-bookworm AS builder

WORKDIR /app

# BuildKit needs this to pass arch info
ARG TARGETARCH

# Copy go mod first to leverage cache
COPY go.mod go.sum ./

# Use BuildKit cache mount for modules
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
COPY . .

# Use BuildKit cache for mod + build cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -o bin/main ./cmd/main.go

# Stage 2: Minimal runtime
FROM gcr.io/distroless/cc

WORKDIR /app

COPY --from=builder /app/bin/main bin/main

EXPOSE 8080

ENTRYPOINT ["./bin/main"]
