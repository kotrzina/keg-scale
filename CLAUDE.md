# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Keg Scale is an IoT project for monitoring beer kegs. An Arduino reads weight from a scale via serial port, sends data to a Go backend over HTTP, which exposes metrics to Prometheus and provides a React frontend dashboard. A WhatsApp bot (Mr. Botka) notifies about significant events.

## Architecture

**Components:**
- `backend/` - Go HTTP server (main application)
- `frontend/` - React dashboard (Create React App)
- `firmware/` - Arduino sketch for Arduino Uno WiFi Rev2

**Backend packages (`backend/pkg/`):**
- `scale/` - Core keg monitoring logic (measurements, warehouse inventory, bank balance)
- `ai/` - AI chat integration (OpenAI + Anthropic) with tool providers for local info
- `wa/` - WhatsApp client using whatsmeow
- `hook/` - WhatsApp bot message handler
- `store/` - Redis storage
- `config/` - Environment configuration
- `prometheus/` - Metrics registry
- `promector/` - Prometheus range query client

**Data flow:** Arduino → POST `/api/scale/push` → Scale service → Prometheus metrics + Redis storage

## Common Commands

### Backend (from `backend/` directory)
```bash
make build        # Build binary
make test         # Run tests with race detector
make lint         # Run golangci-lint
make lint-fix     # Run golangci-lint with auto-fix
make deps         # go mod tidy && verify
```

### Frontend (from `frontend/` directory)
```bash
npm start         # Development server
npm run build     # Production build
npm test          # Run tests
```

### Firmware (from `firmware/` directory)
```bash
make compile      # Compile with arduino-cli
make flush        # Compile and upload to connected Arduino
make deps         # Install required Arduino libraries
make read         # Open serial monitor (screen)
```

### Docker
Build combined container from repo root:
```bash
docker build -t keg-scale .
```

## Development Notes

- **Never run the binary directly** - it may affect production. Write unit tests instead.
- Frontend uses `REACT_APP_BACKEND_PREFIX` env var to point to backend in development
- Backend serves frontend static files from `FRONTEND_PATH` env var
- Backend listens on port 8080
- Go 1.24+, Node 23+
