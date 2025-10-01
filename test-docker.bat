@echo off
echo Testing Docker configuration for WebAIlyzer Lite API...

REM Test Docker build
echo Building Docker image...
docker build -t webailyzer-lite-api-test .

if %errorlevel% neq 0 (
    echo ❌ Docker build failed
    exit /b 1
)
echo ✅ Docker build successful

REM Test Docker run
echo Starting container...
docker run -d --name webailyzer-lite-test -p 8082:8080 webailyzer-lite-api-test

REM Wait for container to start
timeout /t 5 /nobreak >nul

REM Test health endpoint
echo Testing health endpoint...
curl -s -o nul -w "%%{http_code}" http://localhost:8082/health > temp_response.txt
set /p response=<temp_response.txt
del temp_response.txt

if "%response%"=="200" (
    echo ✅ Health endpoint working
) else (
    echo ❌ Health endpoint failed ^(HTTP %response%^)
)

REM Test analyze endpoint
echo Testing analyze endpoint...
curl -s -o nul -w "%%{http_code}" -X POST http://localhost:8082/v1/analyze -H "Content-Type: application/json" -d "{\"url\":\"https://example.com\"}" > temp_response.txt
set /p response=<temp_response.txt
del temp_response.txt

if "%response%"=="200" (
    echo ✅ Analyze endpoint working
) else (
    echo ❌ Analyze endpoint failed ^(HTTP %response%^)
)

REM Cleanup
echo Cleaning up...
docker stop webailyzer-lite-test
docker rm webailyzer-lite-test
docker rmi webailyzer-lite-api-test

echo Docker configuration test complete!