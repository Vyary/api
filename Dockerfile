# Stage 1: Build the binary
FROM golang:1.24.5-bookworm AS builder

# Set up a working directory
WORKDIR /app

# Cache dependencies by copying go.mod and go.sum first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary for multiple architectures
ARG TARGETARCH
RUN CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o bin/main ./cmd/main.go

# Stage 2: Final image
FROM gcr.io/distroless/cc

# Set the working directory
WORKDIR /app

COPY --from=builder /app/bin/main bin/main

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./bin/main"]
