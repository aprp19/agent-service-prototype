#!/bin/bash

# Generate Go code from proto files
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc

set -e

echo "🔧 Generating gRPC code from proto files..."

# Create output directory if it doesn't exist
mkdir -p proto/example

# Generate Go code
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/example/example.proto \
  proto/nuha-auth/auth.proto

echo "✅ gRPC code generated successfully!"
