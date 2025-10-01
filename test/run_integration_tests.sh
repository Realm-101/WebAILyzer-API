#!/bin/bash

# Integration Test Runner Script
# This script runs integration tests for WebAIlyzer Lite API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is available
check_docker() {
    print_status "Checking Docker availability..."
    if ! command -v docker &> /dev/null; then
        print_warning "Docker is not available. Docker tests will be skipped."
        return 1
    fi
    
    if ! docker version > /dev/null 2>&1; then
        print_warning "Docker is not running. Docker tests will be skipped."
        return 1
    fi
    
    print_status "Docker is available"
    return 0
}

# Function to check if Go is available
check_go() {
    print_status "Checking Go installation..."
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    print_status "Go is available: $GO_VERSION"
}

# Function to build the application
build_app() {
    print_status "Building WebAIlyzer Lite API..."
    
    cd "$PROJECT_ROOT"
    
    # Clean any existing builds
    rm -f webailyzer-api-test
    
    # Build the application
    go build -o webailyzer-api-test ./cmd/webailyzer-api
    
    if [ $? -eq 0 ]; then
        print_status "Application built successfully"
    else
        print_error "Failed to build application"
        exit 1
    fi
}

# Function to clean up build artifacts
cleanup_build() {
    print_status "Cleaning up build artifacts..."
    cd "$PROJECT_ROOT"
    rm -f webailyzer-api-test
    rm -f coverage.out coverage.html
    rm -rf profiles
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run integration tests with verbose output
    go test -v ./test/integration/... -timeout=30m
    
    if [ $? -eq 0 ]; then
        print_status "Integration tests passed!"
    else
        print_error "Integration tests failed!"
        exit 1
    fi
}

# Function to run only basic integration tests (no Docker)
run_basic_integration_tests() {
    print_status "Running basic integration tests (excluding Docker)..."
    
    cd "$PROJECT_ROOT"
    
    # Run integration tests excluding Docker tests
    go test -v ./test/integration/... -timeout=20m -short
    
    if [ $? -eq 0 ]; then
        print_status "Basic integration tests passed!"
    else
        print_error "Basic integration tests failed!"
        exit 1
    fi
}

# Function to run Docker integration tests only
run_docker_tests() {
    print_status "Running Docker integration tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run only Docker tests
    go test -v ./test/integration/ -run TestDockerIntegration -timeout=30m
    
    if [ $? -eq 0 ]; then
        print_status "Docker integration tests passed!"
    else
        print_error "Docker integration tests failed!"
        exit 1
    fi
}

# Function to run benchmark tests
run_benchmark_tests() {
    print_status "Running benchmark tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run benchmark tests
    go test -v -bench=. -benchmem ./test/benchmarks/... -timeout=60m
    
    if [ $? -eq 0 ]; then
        print_status "Benchmark tests completed!"
    else
        print_error "Benchmark tests failed!"
        exit 1
    fi
}

# Function to run performance profiling
run_performance_profiling() {
    print_status "Running performance profiling..."
    
    cd "$PROJECT_ROOT"
    
    # Create profiles directory
    mkdir -p profiles
    
    # Run benchmarks with CPU profiling
    go test -bench=. -cpuprofile=profiles/cpu.prof ./test/benchmarks/... -timeout=30m
    
    # Run benchmarks with memory profiling
    go test -bench=. -memprofile=profiles/mem.prof ./test/benchmarks/... -timeout=30m
    
    print_status "Performance profiles saved to ./profiles/"
    print_status "View CPU profile: go tool pprof profiles/cpu.prof"
    print_status "View memory profile: go tool pprof profiles/mem.prof"
}

# Function to generate test coverage report
generate_coverage() {
    print_status "Generating test coverage report..."
    
    cd "$PROJECT_ROOT"
    
    # Run tests with coverage (including integration tests)
    go test -coverprofile=coverage.out ./cmd/webailyzer-api/... ./test/integration/...
    
    if [ -f coverage.out ]; then
        # Generate HTML coverage report
        go tool cover -html=coverage.out -o coverage.html
        
        # Show coverage summary
        go tool cover -func=coverage.out | tail -1
        
        print_status "Coverage report generated: coverage.html"
    else
        print_warning "No coverage data generated"
    fi
}

# Function to run all tests
run_all_tests() {
    print_status "Running all tests (unit + integration + benchmarks)..."
    
    cd "$PROJECT_ROOT"
    
    # Run unit tests first
    print_status "Running unit tests..."
    go test -v ./cmd/webailyzer-api/... -short
    
    # Build application for integration tests
    build_app
    
    # Run integration tests
    run_integration_tests
    
    # Run benchmark tests
    run_benchmark_tests
    
    print_status "All tests completed successfully!"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build       Build the application"
    echo "  basic       Run basic integration tests (no Docker)"
    echo "  integration Run all integration tests"
    echo "  docker      Run Docker integration tests only"
    echo "  benchmarks  Run benchmark tests only"
    echo "  profile     Run performance profiling"
    echo "  coverage    Generate test coverage report"
    echo "  all         Run all tests (unit + integration + benchmarks)"
    echo "  cleanup     Clean up build artifacts and test files"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 basic       # Quick integration tests without Docker"
    echo "  $0 integration # Full integration test suite"
    echo "  $0 docker      # Test Docker container functionality"
    echo "  $0 all         # Complete test suite"
}

# Main script logic
case "${1:-help}" in
    build)
        check_go
        build_app
        ;;
    basic)
        check_go
        build_app
        run_basic_integration_tests
        ;;
    integration)
        check_go
        check_docker
        build_app
        run_integration_tests
        ;;
    docker)
        check_go
        check_docker
        build_app
        run_docker_tests
        ;;
    benchmarks)
        check_go
        build_app
        run_benchmark_tests
        ;;
    profile)
        check_go
        build_app
        run_performance_profiling
        ;;
    coverage)
        check_go
        generate_coverage
        ;;
    all)
        check_go
        check_docker
        run_all_tests
        ;;
    cleanup)
        cleanup_build
        ;;
    help|*)
        show_usage
        ;;
esac