@echo off
REM Integration Test Runner Script for Windows
REM This script runs integration tests for WebAIlyzer Lite API

setlocal enabledelayedexpansion

REM Configuration
set PROJECT_ROOT=%~dp0..
cd /d "%PROJECT_ROOT%"

REM Colors (limited support in Windows)
set GREEN=[92m
set RED=[91m
set YELLOW=[93m
set NC=[0m

REM Function to print status
:print_status
echo %GREEN%[INFO]%NC% %~1
goto :eof

:print_error
echo %RED%[ERROR]%NC% %~1
goto :eof

:print_warning
echo %YELLOW%[WARN]%NC% %~1
goto :eof

REM Function to check if Go is available
:check_go
call :print_status "Checking Go installation..."
go version >nul 2>&1
if errorlevel 1 (
    call :print_error "Go is not installed or not in PATH"
    exit /b 1
)
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
call :print_status "Go is available: %GO_VERSION%"
goto :eof

REM Function to check if Docker is available
:check_docker
call :print_status "Checking Docker availability..."
docker version >nul 2>&1
if errorlevel 1 (
    call :print_warning "Docker is not available. Docker tests will be skipped."
    exit /b 1
)
call :print_status "Docker is available"
goto :eof

REM Function to build the application
:build_app
call :print_status "Building WebAIlyzer Lite API..."

REM Clean any existing builds
if exist webailyzer-api-test.exe del webailyzer-api-test.exe

REM Build the application
go build -o webailyzer-api-test.exe ./cmd/webailyzer-api
if errorlevel 1 (
    call :print_error "Failed to build application"
    exit /b 1
)

call :print_status "Application built successfully"
goto :eof

REM Function to clean up build artifacts
:cleanup_build
call :print_status "Cleaning up build artifacts..."
if exist webailyzer-api-test.exe del webailyzer-api-test.exe
if exist coverage.out del coverage.out
if exist coverage.html del coverage.html
if exist profiles rmdir /s /q profiles
goto :eof

REM Function to run basic integration tests
:run_basic_integration_tests
call :print_status "Running basic integration tests (excluding Docker)..."

go test -v ./test/integration/... -timeout=20m -short
if errorlevel 1 (
    call :print_error "Basic integration tests failed!"
    exit /b 1
)

call :print_status "Basic integration tests passed!"
goto :eof

REM Function to run all integration tests
:run_integration_tests
call :print_status "Running integration tests..."

go test -v ./test/integration/... -timeout=30m
if errorlevel 1 (
    call :print_error "Integration tests failed!"
    exit /b 1
)

call :print_status "Integration tests passed!"
goto :eof

REM Function to run Docker tests only
:run_docker_tests
call :print_status "Running Docker integration tests..."

go test -v ./test/integration/ -run TestDockerIntegration -timeout=30m
if errorlevel 1 (
    call :print_error "Docker integration tests failed!"
    exit /b 1
)

call :print_status "Docker integration tests passed!"
goto :eof

REM Function to run benchmark tests
:run_benchmark_tests
call :print_status "Running benchmark tests..."

go test -v -bench=. -benchmem ./test/benchmarks/... -timeout=60m
if errorlevel 1 (
    call :print_error "Benchmark tests failed!"
    exit /b 1
)

call :print_status "Benchmark tests completed!"
goto :eof

REM Function to generate coverage report
:generate_coverage
call :print_status "Generating test coverage report..."

go test -coverprofile=coverage.out ./cmd/webailyzer-api/... ./test/integration/...
if exist coverage.out (
    go tool cover -html=coverage.out -o coverage.html
    go tool cover -func=coverage.out | findstr "total:"
    call :print_status "Coverage report generated: coverage.html"
) else (
    call :print_warning "No coverage data generated"
)
goto :eof

REM Function to run all tests
:run_all_tests
call :print_status "Running all tests (unit + integration + benchmarks)..."

REM Run unit tests first
call :print_status "Running unit tests..."
go test -v ./cmd/webailyzer-api/... -short
if errorlevel 1 (
    call :print_error "Unit tests failed!"
    exit /b 1
)

REM Build application for integration tests
call :build_app
if errorlevel 1 exit /b 1

REM Run integration tests
call :run_integration_tests
if errorlevel 1 exit /b 1

REM Run benchmark tests
call :run_benchmark_tests
if errorlevel 1 exit /b 1

call :print_status "All tests completed successfully!"
goto :eof

REM Function to show usage
:show_usage
echo Usage: %~nx0 [COMMAND]
echo.
echo Commands:
echo   build       Build the application
echo   basic       Run basic integration tests (no Docker)
echo   integration Run all integration tests
echo   docker      Run Docker integration tests only
echo   benchmarks  Run benchmark tests only
echo   coverage    Generate test coverage report
echo   all         Run all tests (unit + integration + benchmarks)
echo   cleanup     Clean up build artifacts and test files
echo   help        Show this help message
echo.
echo Examples:
echo   %~nx0 basic       # Quick integration tests without Docker
echo   %~nx0 integration # Full integration test suite
echo   %~nx0 docker      # Test Docker container functionality
echo   %~nx0 all         # Complete test suite
goto :eof

REM Main script logic
set COMMAND=%1
if "%COMMAND%"=="" set COMMAND=help

if "%COMMAND%"=="build" (
    call :check_go
    if errorlevel 1 exit /b 1
    call :build_app
) else if "%COMMAND%"=="basic" (
    call :check_go
    if errorlevel 1 exit /b 1
    call :build_app
    if errorlevel 1 exit /b 1
    call :run_basic_integration_tests
) else if "%COMMAND%"=="integration" (
    call :check_go
    if errorlevel 1 exit /b 1
    call :check_docker
    call :build_app
    if errorlevel 1 exit /b 1
    call :run_integration_tests
) else if "%COMMAND%"=="docker" (
    call :check_go
    if errorlevel 1 exit /b 1
    call :check_docker
    if errorlevel 1 exit /b 1
    call :build_app
    if errorlevel 1 exit /b 1
    call :run_docker_tests
) else if "%COMMAND%"=="benchmarks" (
    call :check_go
    if errorlevel 1 exit /b 1
    call :build_app
    if errorlevel 1 exit /b 1
    call :run_benchmark_tests
) else if "%COMMAND%"=="coverage" (
    call :check_go
    if errorlevel 1 exit /b 1
    call :generate_coverage
) else if "%COMMAND%"=="all" (
    call :check_go
    if errorlevel 1 exit /b 1
    call :check_docker
    call :run_all_tests
) else if "%COMMAND%"=="cleanup" (
    call :cleanup_build
) else (
    call :show_usage
)

endlocal