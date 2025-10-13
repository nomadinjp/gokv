#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Define output directory
OUTPUT_DIR="./bin"
mkdir -p $OUTPUT_DIR

# Define build flags for static linking and cross-compilation (for Docker)
# CGO_ENABLED=0 is crucial for static binaries and distroless images
BUILD_FLAGS=("-ldflags" "-s -w")

echo "Building gokv server..."
CGO_ENABLED=0 go build "${BUILD_FLAGS[@]}" -o $OUTPUT_DIR/gokv ./cmd/gokv

echo "Building jwt-gen tool..."
CGO_ENABLED=0 go build "${BUILD_FLAGS[@]}" -o $OUTPUT_DIR/jwt-gen ./cmd/jwt-gen

echo "Build complete. Binaries are in $OUTPUT_DIR/"
