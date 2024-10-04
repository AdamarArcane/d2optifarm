#!/bin/bash

# Set the paths to your backend and frontend directories
BACKEND_DIR="./"
FRONTEND_DIR="./d2of-fe"

# Function to clean up background processes on exit
cleanup() {
    echo "Stopping Vite frontend..."
    kill "$FRONTEND_PID"
    echo "Stopping Go backend..."
    kill "$BACKEND_PID"
    exit 0
}

# Trap SIGINT (Ctrl+C) and call cleanup
trap cleanup SIGINT

# Start the Go backend
echo "Starting Go backend..."
cd "$BACKEND_DIR" || exit
go run main.go &
BACKEND_PID=$!

# Start the Vite frontend
echo "Starting Vite frontend..."
cd "$FRONTEND_DIR" || exit
npm run dev &
FRONTEND_PID=$!

# Wait for both processes to finish
wait "$BACKEND_PID"
wait "$FRONTEND_PID"
