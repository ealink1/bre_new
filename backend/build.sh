#!/bin/bash

# Set target OS to linux (matching the .bat configuration)
# This allows building a Linux binary on macOS.
echo "Building for Linux (amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o bre_new -ldflags "-s -w" main.go

if [ $? -eq 0 ]; then
    echo "Build successful. Binary 'main' created."
else
    echo "Build failed."
    exit 1
fi
