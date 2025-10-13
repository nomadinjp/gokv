# --- Stage 1: Builder ---
FROM golang:1.25-alpine AS builder

# Set necessary environment variables for CGO_ENABLED=0 build
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the server binary
# The server is the main application, named 'gokv'
RUN go build -ldflags "-s -w" -o /gokv ./cmd/gokv

# Build the jwt-gen tool (optional for final image, but good to have it built)
RUN go build -ldflags "-s -w" -o /jwt-gen ./cmd/jwt-gen


# --- Stage 2: Final Image ---
# Use distroless for a minimal, secure final image
FROM gcr.io/distroless/static-debian12

# Set the working directory
WORKDIR /

# Copy the compiled server binary from the builder stage
COPY --from=builder /gokv /gokv

# Expose the default port (8080)
EXPOSE 8080

# Set the entrypoint to run the server
ENTRYPOINT ["/gokv"]
