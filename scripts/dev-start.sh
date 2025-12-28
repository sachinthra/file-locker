#!/bin/bash

# File Locker Development Server Starter
# This script starts both backend and frontend development servers

set -e

echo "🔐 File Locker - Starting Development Environment"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Check if backend services are running
echo "📦 Checking backend services..."
if ! docker ps | grep -q "lock_files"; then
    echo "⚠️  Backend services not running. Starting with docker-compose..."
    docker-compose up -d
    echo "⏳ Waiting for services to be ready..."
    sleep 5
else
    echo "✅ Backend services already running"
fi

# Check backend health
echo "🏥 Checking backend health..."
BACKEND_HEALTH=$(curl -s http://localhost:9010/health || echo "unhealthy")
if [[ "$BACKEND_HEALTH" == *"healthy"* ]]; then
    echo "✅ Backend is healthy"
else
    echo "⚠️  Backend might not be ready yet. Check logs with: docker-compose logs -f"
fi

# Start frontend
echo ""
echo "🎨 Starting frontend development server..."
cd frontend

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "📦 Installing frontend dependencies..."
    npm install
fi

echo ""
echo "✅ Starting frontend at http://localhost:5173"
echo "✅ Backend API at http://localhost:9010"
echo "✅ MinIO Console at http://localhost:9013"
echo ""
echo "Press Ctrl+C to stop the frontend server"
echo "To stop backend: docker-compose down"
echo ""

npm run dev
