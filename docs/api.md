# API Documentation

This document describes the VPS Screener API endpoints and their usage.

## Base URL

```
http://localhost:3001/api/v1
```

## Authentication

All API endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### Metrics

#### POST /metrics

Upload metrics from an agent.

**Request Body:**
```json
{
  "node_id": "string",
  "timestamp": "2024-03-20T12:00:00Z",
  "metrics": {
    "cpu_percent": 45.2,
    "memory_percent": 60.1,
    "disk_usage": 75.3,
    "custom_metrics": {
      "custom_metric_a": 42,
      "custom_status": "ok"
    }
  }
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Metrics uploaded successfully"
}
```

### Tasks

#### GET /tasks

Get pending tasks for a node.

**Query Parameters:**
- `node_id` (required): The ID of the node

**Response:**
```json
{
  "tasks": [
    {
      "id": "string",
      "type": "string",
      "command": "string",
      "parameters": {},
      "created_at": "2024-03-20T12:00:00Z"
    }
  ]
}
```

#### POST /tasks/{id}/result

Submit task execution result.

**Path Parameters:**
- `id` (required): Task ID

**Request Body:**
```json
{
  "status": "success",
  "output": "string",
  "error": "string",
  "execution_time": 1.5
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Task result recorded"
}
```

### Nodes

#### GET /nodes

List all registered nodes.

**Response:**
```json
{
  "nodes": [
    {
      "id": "string",
      "hostname": "string",
      "last_seen": "2024-03-20T12:00:00Z",
      "status": "online",
      "version": "1.0.0"
    }
  ]
}
```

#### GET /nodes/{id}

Get details for a specific node.

**Path Parameters:**
- `id` (required): Node ID

**Response:**
```json
{
  "id": "string",
  "hostname": "string",
  "last_seen": "2024-03-20T12:00:00Z",
  "status": "online",
  "version": "1.0.0",
  "projects": [
    {
      "name": "string",
      "status": "running",
      "metrics": {
        "cpu_percent": 45.2,
        "memory_percent": 60.1
      }
    }
  ]
}
```

### Projects

#### GET /projects

List all projects across all nodes.

**Response:**
```json
{
  "projects": [
    {
      "name": "string",
      "node_id": "string",
      "status": "running",
      "last_updated": "2024-03-20T12:00:00Z"
    }
  ]
}
```

#### GET /projects/{name}

Get details for a specific project.

**Path Parameters:**
- `name` (required): Project name

**Response:**
```json
{
  "name": "string",
  "nodes": [
    {
      "node_id": "string",
      "status": "running",
      "metrics": {
        "cpu_percent": 45.2,
        "memory_percent": 60.1
      }
    }
  ]
}
```

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "Invalid request parameters",
  "details": "string"
}
```

### 401 Unauthorized
```json
{
  "error": "Invalid or missing authentication token"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found",
  "details": "string"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "details": "string"
}
```

## Rate Limiting

API requests are limited to:
- 100 requests per minute per IP
- 1000 requests per hour per IP

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1616236800
```

## WebSocket API

### Connection

Connect to the WebSocket endpoint:
```
ws://localhost:3001/api/v1/ws
```

Include the JWT token as a query parameter:
```
ws://localhost:3001/api/v1/ws?token=<your-jwt-token>
```

### Events

#### Node Status Updates
```json
{
  "type": "node_status",
  "data": {
    "node_id": "string",
    "status": "online",
    "timestamp": "2024-03-20T12:00:00Z"
  }
}
```

#### Project Metrics Updates
```json
{
  "type": "project_metrics",
  "data": {
    "project_name": "string",
    "node_id": "string",
    "metrics": {
      "cpu_percent": 45.2,
      "memory_percent": 60.1
    },
    "timestamp": "2024-03-20T12:00:00Z"
  }
}
```

#### Task Updates
```json
{
  "type": "task_update",
  "data": {
    "task_id": "string",
    "status": "completed",
    "output": "string",
    "timestamp": "2024-03-20T12:00:00Z"
  }
}
```

## SDK Examples

### Node.js
```javascript
const axios = require('axios');

const api = axios.create({
  baseURL: 'http://localhost:3001/api/v1',
  headers: {
    'Authorization': `Bearer ${process.env.API_TOKEN}`
  }
});

async function getNodeMetrics(nodeId) {
  try {
    const response = await api.get(`/nodes/${nodeId}`);
    return response.data;
  } catch (error) {
    console.error('Error fetching node metrics:', error);
    throw error;
  }
}
```

### Python
```python
import requests

class VPSScreenerAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {'Authorization': f'Bearer {token}'}

    def get_node_metrics(self, node_id):
        response = requests.get(
            f'{self.base_url}/nodes/{node_id}',
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

api = VPSScreenerAPI('http://localhost:3001/api/v1', 'your-token')
metrics = api.get_node_metrics('node-1')
```

### Go
```go
package main

import (
    "fmt"
    "net/http"
    "encoding/json"
)

type VPSScreenerAPI struct {
    BaseURL string
    Token   string
    Client  *http.Client
}

func (api *VPSScreenerAPI) GetNodeMetrics(nodeID string) (map[string]interface{}, error) {
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/nodes/%s", api.BaseURL, nodeID), nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Token))

    resp, err := api.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result, nil
}
```

## Next Steps

- [Installation Guide](installation.md)
- [Plugin Development Guide](plugin-development.md)
- [Contributing Guide](../CONTRIBUTING.md) 