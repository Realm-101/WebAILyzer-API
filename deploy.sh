#!/bin/bash

# WebAIlyzer Lite API - Simple Deployment Script
# This script provides an easy way to deploy the API using Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
IMAGE_NAME="webailyzer-lite-api"
CONTAINER_NAME="webailyzer-api"
PORT="8080"
MEMORY_LIMIT="256m"
CPU_LIMIT="0.5"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is installed and running
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    
    print_success "Docker is available and running"
}

# Function to build the Docker image
build_image() {
    print_status "Building Docker image: $IMAGE_NAME"
    
    if docker build -t "$IMAGE_NAME" .; then
        print_success "Docker image built successfully"
    else
        print_error "Failed to build Docker image"
        exit 1
    fi
}

# Function to stop and remove existing container
cleanup_existing() {
    if docker ps -a --format 'table {{.Names}}' | grep -q "^$CONTAINER_NAME$"; then
        print_warning "Stopping and removing existing container: $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" &> /dev/null || true
        docker rm "$CONTAINER_NAME" &> /dev/null || true
        print_success "Existing container removed"
    fi
}

# Function to run the container
run_container() {
    print_status "Starting container: $CONTAINER_NAME"
    
    docker run -d \
        --name "$CONTAINER_NAME" \
        -p "$PORT:8080" \
        --memory="$MEMORY_LIMIT" \
        --cpus="$CPU_LIMIT" \
        --restart unless-stopped \
        --read-only \
        --tmpfs /tmp \
        "$IMAGE_NAME"
    
    if [ $? -eq 0 ]; then
        print_success "Container started successfully"
        print_status "Container is running on port $PORT"
    else
        print_error "Failed to start container"
        exit 1
    fi
}

# Function to wait for the service to be ready
wait_for_service() {
    print_status "Waiting for service to be ready..."
    
    for i in {1..30}; do
        if curl -s -f "http://localhost:$PORT/health" &> /dev/null; then
            print_success "Service is ready and responding"
            return 0
        fi
        sleep 1
    done
    
    print_error "Service did not become ready within 30 seconds"
    print_status "Check container logs with: docker logs $CONTAINER_NAME"
    return 1
}

# Function to test the API
test_api() {
    print_status "Testing API endpoints..."
    
    # Test health endpoint
    if curl -s -f "http://localhost:$PORT/health" | grep -q '"status":"ok"'; then
        print_success "Health endpoint is working"
    else
        print_error "Health endpoint test failed"
        return 1
    fi
    
    # Test analysis endpoint with a simple URL
    if curl -s -X POST "http://localhost:$PORT/v1/analyze" \
        -H "Content-Type: application/json" \
        -d '{"url":"https://httpbin.org/html"}' | grep -q '"detected"'; then
        print_success "Analysis endpoint is working"
    else
        print_warning "Analysis endpoint test failed (this might be due to network issues)"
    fi
}

# Function to show usage information
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -p, --port PORT        Port to expose (default: 8080)"
    echo "  -m, --memory LIMIT     Memory limit (default: 256m)"
    echo "  -c, --cpu LIMIT        CPU limit (default: 0.5)"
    echo "  -n, --name NAME        Container name (default: webailyzer-api)"
    echo "  -i, --image NAME       Image name (default: webailyzer-lite-api)"
    echo "  --no-test             Skip API testing after deployment"
    echo "  --cleanup-only        Only cleanup existing container and exit"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Deploy with default settings"
    echo "  $0 -p 9000           # Deploy on port 9000"
    echo "  $0 -m 512m -c 1.0    # Deploy with more resources"
    echo "  $0 --cleanup-only    # Only remove existing container"
}

# Function to show deployment information
show_deployment_info() {
    echo ""
    print_success "=== Deployment Complete ==="
    echo ""
    echo "Service Information:"
    echo "  • Container Name: $CONTAINER_NAME"
    echo "  • Image: $IMAGE_NAME"
    echo "  • Port: $PORT"
    echo "  • Memory Limit: $MEMORY_LIMIT"
    echo "  • CPU Limit: $CPU_LIMIT"
    echo ""
    echo "API Endpoints:"
    echo "  • Health Check: http://localhost:$PORT/health"
    echo "  • Analysis: http://localhost:$PORT/v1/analyze"
    echo ""
    echo "Quick Test Commands:"
    echo "  curl http://localhost:$PORT/health"
    echo "  curl -X POST http://localhost:$PORT/v1/analyze \\"
    echo "    -H 'Content-Type: application/json' \\"
    echo "    -d '{\"url\":\"https://example.com\"}'"
    echo ""
    echo "Management Commands:"
    echo "  • View logs: docker logs $CONTAINER_NAME"
    echo "  • Stop service: docker stop $CONTAINER_NAME"
    echo "  • Start service: docker start $CONTAINER_NAME"
    echo "  • Remove service: docker rm -f $CONTAINER_NAME"
    echo ""
}

# Parse command line arguments
SKIP_TEST=false
CLEANUP_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -m|--memory)
            MEMORY_LIMIT="$2"
            shift 2
            ;;
        -c|--cpu)
            CPU_LIMIT="$2"
            shift 2
            ;;
        -n|--name)
            CONTAINER_NAME="$2"
            shift 2
            ;;
        -i|--image)
            IMAGE_NAME="$2"
            shift 2
            ;;
        --no-test)
            SKIP_TEST=true
            shift
            ;;
        --cleanup-only)
            CLEANUP_ONLY=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main deployment process
main() {
    echo "WebAIlyzer Lite API - Deployment Script"
    echo "========================================"
    echo ""
    
    # Check prerequisites
    check_docker
    
    # Cleanup existing container
    cleanup_existing
    
    # If cleanup-only mode, exit here
    if [ "$CLEANUP_ONLY" = true ]; then
        print_success "Cleanup completed"
        exit 0
    fi
    
    # Build and deploy
    build_image
    run_container
    
    # Wait for service to be ready
    if wait_for_service; then
        # Test the API if not skipped
        if [ "$SKIP_TEST" = false ]; then
            test_api
        fi
        
        # Show deployment information
        show_deployment_info
    else
        print_error "Deployment failed - service is not responding"
        print_status "Check container logs with: docker logs $CONTAINER_NAME"
        exit 1
    fi
}

# Run main function
main "$@"