#!/bin/bash

# Builds all language runtime images

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Building from: $PROJECT_DIR"
echo ""

echo "Building python image"
docker build -t juri/python "$PROJECT_DIR/docker/python/"

echo "Building cpp image..."
docker build -t juri/cpp "$PROJECT_DIR/docker/cpp"

echo "Building c image..."
docker build -t juri/c "$PROJECT_DIR/docker/c"


echo "Building java image..."
docker build -t juri/java "$PROJECT_DIR/docker/java/"


echo ""
echo "All images built successfully!"
echo "Images:"
docker images | grep juri