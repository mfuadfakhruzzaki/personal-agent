# AI-Powered Todo Agent Backend

Backend service dalam Go untuk mengonversi teks, gambar, atau dokumen menjadi daftar tugas terstruktur menggunakan Gemini AI.

## Features

- 🚀 Lightweight Go backend optimized for AWS EC2 Free Tier
- 🤖 AI-powered task extraction using Gemini API
- 📄 Support for text, image, and document inputs
- 🗄️ Supabase integration for data persistence
- 🔄 Asynchronous processing with job queue
- 📊 Structured logging and monitoring
- 🐳 Docker containerized deployment

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Supabase account
- Google Gemini API key

### Installation

1. Clone the repository:

```bash
git clone <repo-url>
cd todo-agent-backend
```

2. Copy environment config:

```bash
cp config/config.example.yaml config/config.yaml
```

3. Set your environment variables:

```bash
export GEMINI_API_KEY="your-gemini-api-key"
export SUPABASE_URL="your-supabase-url"
export SUPABASE_KEY="your-supabase-anon-key"
```

4. Run with Docker Compose:

```bash
docker-compose up -d
```

### Development

```bash
# Install dependencies
go mod tidy

# Run locally
go run cmd/server/main.go

# Run tests
go test ./...

# Build binary
go build -o bin/todo-agent cmd/server/main.go
```

## API Documentation

### POST /process

Process text, image, or document input to extract todo items.

**Request:**

```bash
curl -X POST http://localhost:8080/process \
  -H "X-API-Key: your-api-key" \
  -F "type=text" \
  -F "content=Besok meeting dengan client jam 10, lalu review code, dan kirim laporan ke manager" \
  -F "user_id=user123"
```

**Response:**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "accepted",
  "message": "Processing request"
}
```

### GET /status/{job_id}

Check processing status and results.

**Response:**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "todos": [
    {
      "title": "Meeting dengan client",
      "description": "Meeting scheduled with client",
      "due_date": "2025-07-16"
    }
  ]
}
```

### GET /healthz

Health check endpoint.

## Configuration

See `config/config.yaml` for all configuration options:

- Server settings (port, timeouts)
- Database connection
- API keys and external services
- Logging configuration
- File upload limits

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Frontend      │────│   Go Backend     │────│   Supabase      │
│   (Next.js)     │    │   (Gin/Fiber)    │    │   (PostgreSQL)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                       ┌──────────────────┐
                       │   Gemini API     │
                       │   (Google AI)    │
                       └──────────────────┘
```

## Performance

- Memory usage: ~50MB idle, <400MB peak
- Response time: <2s P95 latency
- Throughput: 5 requests/second per IP
- Uptime target: 99%

## Deployment

### AWS EC2 Free Tier

1. Launch t2.micro/t3.micro instance
2. Install Docker
3. Clone repository
4. Set environment variables
5. Run with docker-compose

### Systemd Service

```bash
# Copy service file
sudo cp deploy/todo-agent.service /etc/systemd/system/

# Enable and start
sudo systemctl enable todo-agent
sudo systemctl start todo-agent
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
