# Todo Agent Backend API Documentation

## Overview

The Todo Agent Backend is a Go-based REST API that converts text, image, or document inputs into structured todo lists using AI processing via Google's Gemini API.

## Base URL

```
http://localhost:8080
```

## Authentication

All API endpoints (except health check) require authentication using an API key.

### Headers

```http
X-API-Key: your-api-key-here
```

Alternatively, you can use the Authorization header:

```http
Authorization: Bearer your-api-key-here
```

## Endpoints

### 1. Health Check

Check if the service is running and healthy.

**Endpoint:** `GET /healthz`

**Authentication:** Not required

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2025-07-15T10:30:00Z",
  "version": "1.0.0"
}
```

**Status Codes:**

- `200 OK` - Service is healthy

---

### 2. Process Input

Submit text, image, or document for todo extraction.

**Endpoint:** `POST /process`

**Authentication:** Required

**Content-Type:** `multipart/form-data`

**Parameters:**

| Field     | Type   | Required    | Description                                      |
| --------- | ------ | ----------- | ------------------------------------------------ |
| `type`    | string | Yes         | Input type: `text`, `image`, or `document`       |
| `user_id` | string | Yes         | Unique identifier for the user                   |
| `content` | string | Conditional | Text content (required if type=text)             |
| `file`    | file   | Conditional | File upload (required if type=image or document) |

**Example Request (Text):**

```bash
curl -X POST http://localhost:8080/process \
  -H "X-API-Key: your-api-key" \
  -F "type=text" \
  -F "content=Meeting tomorrow at 10am, review code, send report by Friday" \
  -F "user_id=user123"
```

**Example Request (Image):**

```bash
curl -X POST http://localhost:8080/process \
  -H "X-API-Key: your-api-key" \
  -F "type=image" \
  -F "file=@notes.jpg" \
  -F "user_id=user123"
```

**Example Request (Document):**

```bash
curl -X POST http://localhost:8080/process \
  -H "X-API-Key: your-api-key" \
  -F "type=document" \
  -F "file=@meeting_notes.pdf" \
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

**Status Codes:**

- `202 Accepted` - Request accepted for processing
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Invalid or missing API key
- `413 Request Entity Too Large` - File too large (max 5MB)
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

**File Restrictions:**

- Maximum file size: 5MB
- Supported image formats: jpg, jpeg, png, gif, bmp, webp
- Supported document formats: pdf, doc, docx, txt, rtf

---

### 3. Get Job Status

Check the processing status and results of a submitted job.

**Endpoint:** `GET /status/{job_id}`

**Authentication:** Required

**Path Parameters:**

| Parameter | Type   | Description                                   |
| --------- | ------ | --------------------------------------------- |
| `job_id`  | string | The job ID returned from the process endpoint |

**Example Request:**

```bash
curl -X GET http://localhost:8080/status/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: your-api-key"
```

**Response (Pending):**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "created_at": "2025-07-15T10:30:00Z",
  "updated_at": "2025-07-15T10:30:00Z"
}
```

**Response (Processing):**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "created_at": "2025-07-15T10:30:00Z",
  "updated_at": "2025-07-15T10:30:05Z"
}
```

**Response (Completed):**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "created_at": "2025-07-15T10:30:00Z",
  "updated_at": "2025-07-15T10:30:10Z",
  "todos": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "user_id": "user123",
      "title": "Meeting with client",
      "description": "Scheduled meeting tomorrow",
      "due_date": "2025-07-16T10:00:00Z",
      "source_type": "text",
      "created_at": "2025-07-15T10:30:10Z"
    },
    {
      "id": "123e4567-e89b-12d3-a456-426614174001",
      "user_id": "user123",
      "title": "Review code",
      "description": "Code review task",
      "due_date": null,
      "source_type": "text",
      "created_at": "2025-07-15T10:30:10Z"
    }
  ]
}
```

**Response (Failed):**

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "failed",
  "message": "Failed to process with AI: API quota exceeded",
  "created_at": "2025-07-15T10:30:00Z",
  "updated_at": "2025-07-15T10:30:15Z"
}
```

**Status Codes:**

- `200 OK` - Job status retrieved successfully
- `401 Unauthorized` - Invalid or missing API key
- `404 Not Found` - Job not found
- `500 Internal Server Error` - Server error

**Job Status Values:**

- `pending` - Job submitted, waiting for processing
- `processing` - Job is currently being processed
- `completed` - Job completed successfully
- `failed` - Job failed with error

---

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Limit:** 5 requests per second per IP address
- **Burst:** 10 requests allowed in burst
- **Response:** `429 Too Many Requests` when limit exceeded

## Error Handling

All errors follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "code": 400
}
```

**Common Error Codes:**

| Code                  | Error | Description                |
| --------------------- | ----- | -------------------------- |
| `validation_error`    | 400   | Invalid request parameters |
| `unauthorized`        | 401   | Invalid or missing API key |
| `not_found`           | 404   | Resource not found         |
| `rate_limit_exceeded` | 429   | Too many requests          |
| `internal_error`      | 500   | Server error               |

## Data Persistence

Extracted todos are automatically saved to the Supabase database with the following schema:

```sql
CREATE TABLE todos (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id text NOT NULL,
    title text NOT NULL,
    description text,
    due_date timestamptz,
    source_type text NOT NULL,
    source_url text,
    created_at timestamptz DEFAULT now()
);
```

## Examples

### Complete Workflow Example

1. **Submit a text for processing:**

```bash
curl -X POST http://localhost:8080/process \
  -H "X-API-Key: your-api-key" \
  -F "type=text" \
  -F "content=Tomorrow: meeting at 9am, lunch with team at 12pm, finish project report by 5pm" \
  -F "user_id=user123"
```

Response:

```json
{
  "job_id": "abc123",
  "status": "accepted",
  "message": "Processing request"
}
```

2. **Check processing status:**

```bash
curl -X GET http://localhost:8080/status/abc123 \
  -H "X-API-Key: your-api-key"
```

3. **Get completed results:**

```json
{
  "job_id": "abc123",
  "status": "completed",
  "todos": [
    {
      "title": "Meeting",
      "description": "Scheduled meeting",
      "due_date": "2025-07-16T09:00:00Z"
    },
    {
      "title": "Lunch with team",
      "description": "Team lunch",
      "due_date": "2025-07-16T12:00:00Z"
    },
    {
      "title": "Finish project report",
      "description": "Complete and submit project report",
      "due_date": "2025-07-16T17:00:00Z"
    }
  ]
}
```

## SDKs and Client Libraries

Currently, the API can be consumed using standard HTTP libraries. Example implementations:

### JavaScript/Node.js

```javascript
const axios = require("axios");
const FormData = require("form-data");

const client = {
  baseURL: "http://localhost:8080",
  apiKey: "your-api-key",

  async processText(content, userId) {
    const form = new FormData();
    form.append("type", "text");
    form.append("content", content);
    form.append("user_id", userId);

    const response = await axios.post(`${this.baseURL}/process`, form, {
      headers: {
        "X-API-Key": this.apiKey,
        ...form.getHeaders(),
      },
    });

    return response.data;
  },

  async getJobStatus(jobId) {
    const response = await axios.get(`${this.baseURL}/status/${jobId}`, {
      headers: {
        "X-API-Key": this.apiKey,
      },
    });

    return response.data;
  },
};
```

### Python

```python
import requests

class TodoAgentClient:
    def __init__(self, base_url="http://localhost:8080", api_key=""):
        self.base_url = base_url
        self.api_key = api_key

    def process_text(self, content, user_id):
        headers = {"X-API-Key": self.api_key}
        data = {
            "type": "text",
            "content": content,
            "user_id": user_id
        }

        response = requests.post(f"{self.base_url}/process",
                               data=data, headers=headers)
        return response.json()

    def get_job_status(self, job_id):
        headers = {"X-API-Key": self.api_key}
        response = requests.get(f"{self.base_url}/status/{job_id}",
                              headers=headers)
        return response.json()
```

## Performance Considerations

- **Memory Usage:** Optimized for AWS EC2 Free Tier (1GB RAM)
- **Response Time:** Target P95 latency < 2 seconds
- **File Processing:** Large files are processed asynchronously
- **Cleanup:** Temporary files are automatically cleaned up after processing

## Monitoring and Health

- **Health Endpoint:** `/healthz` for load balancer health checks
- **Logging:** Structured JSON logs for monitoring
- **Metrics:** Response times and error rates logged
- **Uptime:** Target 99% availability
