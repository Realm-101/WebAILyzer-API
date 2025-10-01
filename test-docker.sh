#!/bin/bash

echo "Testing Docker configuration for WebAIlyzer Lite API..."

# Test Docker build
echo "Building Docker image..."
docker build -t webailyzer-lite-api-test .

if [ $? -eq 0 ]; then
    echo "✅ Docker build successful"
else
    echo "❌ Docker build failed"
    exit 1
fi

# Test Docker run
echo "Starting container..."
docker run -d --name webailyzer-lite-test -p 8082:8080 webailyzer-lite-api-test

# Wait for container to start
sleep 5

# Test health endpoint
echo "Testing health endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/health)

if [ "$response" = "200" ]; then
    echo "✅ Health endpoint working"
else
    echo "❌ Health endpoint failed (HTTP $response)"
fi

# Test analyze endpoint
echo "Testing analyze endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8082/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}')

if [ "$response" = "200" ]; then
    echo "✅ Analyze endpoint working"
else
    echo "❌ Analyze endpoint failed (HTTP $response)"
fi

# Cleanup
echo "Cleaning up..."
docker stop webailyzer-lite-test
docker rm webailyzer-lite-test
docker rmi webailyzer-lite-api-test

echo "Docker configuration test complete!"