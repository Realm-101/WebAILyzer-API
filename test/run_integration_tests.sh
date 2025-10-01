#!/bin/bash

# Integration Test Runner Script
# This script sets up the test environment and runs integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
TEST_DB_NAME="webailyzer_test"
BENCHMARK_DB_NAME="webailyzer_benchmark"
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="password"
POSTGRES_HOST="localhost"
POSTGRES_PORT="5432"

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

# Function to check if PostgreSQL is running
check_postgres() {
    print_status "Checking PostgreSQL connection..."
    if ! pg_isready -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER > /dev/null 2>&1; then
        print_error "PostgreSQL is not running or not accessible"
        print_error "Please ensure PostgreSQL is running on $POSTGRES_HOST:$POSTGRES_PORT"
        exit 1
    fi
    print_status "PostgreSQL is running"
}

# Function to check if Redis is running (optional)
check_redis() {
    print_status "Checking Redis connection..."
    if ! redis-cli ping > /dev/null 2>&1; then
        print_warning "Redis is not running. Some tests may be skipped."
        return 1
    fi
    print_status "Redis is running"
    return 0
}

# Function to create test databases
setup_databases() {
    print_status "Setting up test databases..."
    
    # Create test database
    PGPASSWORD=$POSTGRES_PASSWORD createdb -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER $TEST_DB_NAME 2>/dev/null || true
    
    # Create benchmark database
    PGPASSWORD=$POSTGRES_PASSWORD createdb -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER $BENCHMARK_DB_NAME 2>/dev/null || true
    
    print_status "Test databases created/verified"
}

# Function to clean up test databases
cleanup_databases() {
    print_status "Cleaning up test databases..."
    
    # Drop test database
    PGPASSWORD=$POSTGRES_PASSWORD dropdb -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER $TEST_DB_NAME 2>/dev/null || true
    
    # Drop benchmark database
    PGPASSWORD=$POSTGRES_PASSWORD dropdb -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER $BENCHMARK_DB_NAME 2>/dev/null || true
    
    print_status "Test databases cleaned up"
}

# Function to run integration tests
run_integration_tests() {
    print_status "Running integration tests..."
    
    export POSTGRES_HOST=$POSTGRES_HOST
    export POSTGRES_PORT=$POSTGRES_PORT
    export POSTGRES_USER=$POSTGRES_USER
    export POSTGRES_PASSWORD=$POSTGRES_PASSWORD
    export TEST_DB_NAME=$TEST_DB_NAME
    
    # Run integration tests with verbose output
    go test -v -tags=integration ./test/integration/... -timeout=30m
    
    if [ $? -eq 0 ]; then
        print_status "Integration tests passed!"
    else
        print_error "Integration tests failed!"
        exit 1
    fi
}

# Function to run benchmark tests
run_benchmark_tests() {
    print_status "Running benchmark tests..."
    
    export POSTGRES_HOST=$POSTGRES_HOST
    export POSTGRES_PORT=$POSTGRES_PORT
    export POSTGRES_USER=$POSTGRES_USER
    export POSTGRES_PASSWORD=$POSTGRES_PASSWORD
    export BENCHMARK_DB_NAME=$BENCHMARK_DB_NAME
    
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
    
    # Create profiles directory
    mkdir -p profiles
    
    # Run benchmarks with CPU profiling
    go test -bench=BenchmarkAnalysisEndpoint -cpuprofile=profiles/cpu.prof ./test/benchmarks/...
    
    # Run benchmarks with memory profiling
    go test -bench=BenchmarkMemoryUsage -memprofile=profiles/mem.prof ./test/benchmarks/...
    
    print_status "Performance profiles saved to ./profiles/"
    print_status "View CPU profile: go tool pprof profiles/cpu.prof"
    print_status "View memory profile: go tool pprof profiles/mem.prof"
}

# Function to generate test coverage report
generate_coverage() {
    print_status "Generating test coverage report..."
    
    # Run tests with coverage
    go test -coverprofile=coverage.out ./internal/...
    
    # Generate HTML coverage report
    go tool cover -html=coverage.out -o coverage.html
    
    # Show coverage summary
    go tool cover -func=coverage.out | tail -1
    
    print_status "Coverage report generated: coverage.html"
}

# Function to run all tests
run_all_tests() {
    print_status "Running all tests (unit + integration + benchmarks)..."
    
    # Run unit tests first
    print_status "Running unit tests..."
    go test -v ./internal/... -short
    
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
    echo "  setup       Set up test environment (databases)"
    echo "  integration Run integration tests only"
    echo "  benchmarks  Run benchmark tests only"
    echo "  profile     Run performance profiling"
    echo "  coverage    Generate test coverage report"
    echo "  all         Run all tests (unit + integration + benchmarks)"
    echo "  cleanup     Clean up test environment"
    echo "  help        Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  POSTGRES_HOST     PostgreSQL host (default: localhost)"
    echo "  POSTGRES_PORT     PostgreSQL port (default: 5432)"
    echo "  POSTGRES_USER     PostgreSQL user (default: postgres)"
    echo "  POSTGRES_PASSWORD PostgreSQL password (default: password)"
}

# Main script logic
case "${1:-help}" in
    setup)
        check_postgres
        setup_databases
        ;;
    integration)
        check_postgres
        setup_databases
        run_integration_tests
        ;;
    benchmarks)
        check_postgres
        setup_databases
        run_benchmark_tests
        ;;
    profile)
        check_postgres
        setup_databases
        run_performance_profiling
        ;;
    coverage)
        generate_coverage
        ;;
    all)
        check_postgres
        check_redis
        setup_databases
        run_all_tests
        ;;
    cleanup)
        cleanup_databases
        ;;
    help|*)
        show_usage
        ;;
esac