@echo off
setlocal enabledelayedexpansion

REM WebAIlyzer Lite API - Simple Deployment Script for Windows
REM This script provides an easy way to deploy the API using Docker

REM Default values
set IMAGE_NAME=webailyzer-lite-api
set CONTAINER_NAME=webailyzer-api
set PORT=8080
set MEMORY_LIMIT=256m
set CPU_LIMIT=0.5
set SKIP_TEST=false
set CLEANUP_ONLY=false

REM Parse command line arguments
:parse_args
if "%~1"=="" goto start_deployment
if "%~1"=="-p" (
    set PORT=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--port" (
    set PORT=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="-m" (
    set MEMORY_LIMIT=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--memory" (
    set MEMORY_LIMIT=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="-c" (
    set CPU_LIMIT=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--cpu" (
    set CPU_LIMIT=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="-n" (
    set CONTAINER_NAME=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--name" (
    set CONTAINER_NAME=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="-i" (
    set IMAGE_NAME=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--image" (
    set IMAGE_NAME=%~2
    shift
    shift
    goto parse_args
)
if "%~1"=="--no-test" (
    set SKIP_TEST=true
    shift
    goto parse_args
)
if "%~1"=="--cleanup-only" (
    set CLEANUP_ONLY=true
    shift
    goto parse_args
)
if "%~1"=="-h" goto show_usage
if "%~1"=="--help" goto show_usage

echo [ERROR] Unknown option: %~1
goto show_usage

:show_usage
echo Usage: %~nx0 [OPTIONS]
echo.
echo Options:
echo   -p, --port PORT        Port to expose (default: 8080)
echo   -m, --memory LIMIT     Memory limit (default: 256m)
echo   -c, --cpu LIMIT        CPU limit (default: 0.5)
echo   -n, --name NAME        Container name (default: webailyzer-api)
echo   -i, --image NAME       Image name (default: webailyzer-lite-api)
echo   --no-test             Skip API testing after deployment
echo   --cleanup-only        Only cleanup existing container and exit
echo   -h, --help            Show this help message
echo.
echo Examples:
echo   %~nx0                    # Deploy with default settings
echo   %~nx0 -p 9000           # Deploy on port 9000
echo   %~nx0 -m 512m -c 1.0    # Deploy with more resources
echo   %~nx0 --cleanup-only    # Only remove existing container
goto end

:start_deployment
echo WebAIlyzer Lite API - Deployment Script
echo ========================================
echo.

REM Check if Docker is installed and running
echo [INFO] Checking Docker availability...
docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed. Please install Docker first.
    goto error_exit
)

docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not running. Please start Docker first.
    goto error_exit
)

echo [SUCCESS] Docker is available and running

REM Cleanup existing container
echo [INFO] Checking for existing container: %CONTAINER_NAME%
docker ps -a --format "table {{.Names}}" | findstr /r "^%CONTAINER_NAME%$" >nul 2>&1
if not errorlevel 1 (
    echo [WARNING] Stopping and removing existing container: %CONTAINER_NAME%
    docker stop %CONTAINER_NAME% >nul 2>&1
    docker rm %CONTAINER_NAME% >nul 2>&1
    echo [SUCCESS] Existing container removed
)

REM If cleanup-only mode, exit here
if "%CLEANUP_ONLY%"=="true" (
    echo [SUCCESS] Cleanup completed
    goto end
)

REM Build Docker image
echo [INFO] Building Docker image: %IMAGE_NAME%
docker build -t %IMAGE_NAME% .
if errorlevel 1 (
    echo [ERROR] Failed to build Docker image
    goto error_exit
)
echo [SUCCESS] Docker image built successfully

REM Run container
echo [INFO] Starting container: %CONTAINER_NAME%
docker run -d --name %CONTAINER_NAME% -p %PORT%:8080 --memory=%MEMORY_LIMIT% --cpus=%CPU_LIMIT% --restart unless-stopped --read-only --tmpfs /tmp %IMAGE_NAME%
if errorlevel 1 (
    echo [ERROR] Failed to start container
    goto error_exit
)
echo [SUCCESS] Container started successfully
echo [INFO] Container is running on port %PORT%

REM Wait for service to be ready
echo [INFO] Waiting for service to be ready...
set /a counter=0
:wait_loop
if %counter% geq 30 (
    echo [ERROR] Service did not become ready within 30 seconds
    echo [INFO] Check container logs with: docker logs %CONTAINER_NAME%
    goto error_exit
)
curl -s -f "http://localhost:%PORT%/health" >nul 2>&1
if not errorlevel 1 (
    echo [SUCCESS] Service is ready and responding
    goto test_api
)
timeout /t 1 /nobreak >nul
set /a counter+=1
goto wait_loop

:test_api
if "%SKIP_TEST%"=="true" goto show_info

echo [INFO] Testing API endpoints...

REM Test health endpoint
curl -s -f "http://localhost:%PORT%/health" | findstr "status.*ok" >nul 2>&1
if not errorlevel 1 (
    echo [SUCCESS] Health endpoint is working
) else (
    echo [ERROR] Health endpoint test failed
    goto error_exit
)

REM Test analysis endpoint
curl -s -X POST "http://localhost:%PORT%/v1/analyze" -H "Content-Type: application/json" -d "{\"url\":\"https://httpbin.org/html\"}" | findstr "detected" >nul 2>&1
if not errorlevel 1 (
    echo [SUCCESS] Analysis endpoint is working
) else (
    echo [WARNING] Analysis endpoint test failed (this might be due to network issues)
)

:show_info
echo.
echo [SUCCESS] === Deployment Complete ===
echo.
echo Service Information:
echo   • Container Name: %CONTAINER_NAME%
echo   • Image: %IMAGE_NAME%
echo   • Port: %PORT%
echo   • Memory Limit: %MEMORY_LIMIT%
echo   • CPU Limit: %CPU_LIMIT%
echo.
echo API Endpoints:
echo   • Health Check: http://localhost:%PORT%/health
echo   • Analysis: http://localhost:%PORT%/v1/analyze
echo.
echo Quick Test Commands:
echo   curl http://localhost:%PORT%/health
echo   curl -X POST http://localhost:%PORT%/v1/analyze ^
echo     -H "Content-Type: application/json" ^
echo     -d "{\"url\":\"https://example.com\"}"
echo.
echo Management Commands:
echo   • View logs: docker logs %CONTAINER_NAME%
echo   • Stop service: docker stop %CONTAINER_NAME%
echo   • Start service: docker start %CONTAINER_NAME%
echo   • Remove service: docker rm -f %CONTAINER_NAME%
echo.
goto end

:error_exit
exit /b 1

:end
endlocal