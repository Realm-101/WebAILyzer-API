# WebAIlyzer Lite API - Quick Start Guide

## 🚀 One-Command Deployment

### Linux/Mac
```bash
git clone https://github.com/your-username/webailyzer-lite-api.git
cd webailyzer-lite-api && ./deploy.sh
```

### Windows
```cmd
git clone https://github.com/your-username/webailyzer-lite-api.git
cd webailyzer-lite-api && deploy.bat
```

## 📋 Prerequisites

- Docker 20.10+
- Git

## 🔗 API Endpoints

Once deployed, the API will be available at `http://localhost:8080`:

### Health Check
```bash
curl http://localhost:8080/health
```

### Analyze Website
```bash
curl -X POST http://localhost:8080/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
```

## 🛠️ Management Commands

```bash
# View logs
docker logs webailyzer-api

# Stop service
docker stop webailyzer-api

# Start service
docker start webailyzer-api

# Remove service
docker rm -f webailyzer-api
```

## 📚 Full Documentation

- [Complete Deployment Guide](DEPLOYMENT.md) - Detailed deployment instructions
- [API Documentation](API_DOCUMENTATION.md) - Complete API reference
- [README](README.md) - Project overview and features

## 🆘 Quick Troubleshooting

### Service won't start?
```bash
# Check Docker is running
docker info

# Check logs
docker logs webailyzer-api

# Rebuild and restart
./deploy.sh --cleanup-only && ./deploy.sh
```

### API not responding?
```bash
# Test health endpoint
curl -v http://localhost:8080/health

# Check if port is in use
netstat -tlnp | grep :8080  # Linux
netstat -an | findstr :8080  # Windows
```

### Need help?
Check the [Troubleshooting section](DEPLOYMENT.md#troubleshooting) in the deployment guide.