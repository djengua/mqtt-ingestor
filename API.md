# API v1 Documentation

## Overview
This is the MQTT Ingestor API v1. All endpoints are prefixed with `/api/v1`.

## Authentication

### Register
**POST** `/api/v1/auth/register`

Create a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secure_password",
  "name": "John Doe"
}
```

**Response:** (201 Created)
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

### Login
**POST** `/api/v1/auth/login`

Authenticate a user and get a JWT token.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:** (200 OK)
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

## Protected Endpoints

All protected endpoints require the `Authorization` header with a Bearer token:

```
Authorization: Bearer <your_jwt_token>
```

### List Devices
**GET** `/api/v1/devices`

Get a paginated list of all devices.

**Query Parameters:**
- `limit` (optional, default: 50, max: 500) - Number of devices to return
- `offset` (optional, default: 0) - Number of devices to skip

**Response:** (200 OK)
```json
{
  "devices": [
    {
      "id": 1,
      "device_id": "device_001",
      "node_type": "sensor",
      "gateway_id": "gateway_1",
      "last_seen_at": "2026-04-13T10:30:00Z",
      "last_seq": 12345,
      "last_status": 1,
      "last_battery_v": 4.2,
      "created_at": "2026-04-10T08:00:00Z",
      "updated_at": "2026-04-13T10:30:00Z"
    }
  ],
  "total": 150
}
```

### Get Device Telemetry
**GET** `/api/v1/devices/{device_id}/telemetry`

Get telemetry events for a specific device.

**Path Parameters:**
- `device_id` - The device ID

**Query Parameters:**
- `limit` (optional, default: 100, max: 1000) - Number of events to return
- `offset` (optional, default: 0) - Number of events to skip

**Response:** (200 OK)
```json
{
  "events": [
    {
      "id": 1,
      "device_id": "device_001",
      "topic": "sensores/device_001/telemetry",
      "seq": 12345,
      "node_type": "sensor",
      "gateway_id": "gateway_1",
      "status": 1,
      "temperature_c": 23.5,
      "humidity_air_pct": 65.2,
      "soil_moisture_raw": 512,
      "soil_moisture_pct": 45.0,
      "battery_v": 4.2,
      "device_ts": "2026-04-13T10:30:00Z",
      "ingested_at": "2026-04-13T10:30:01Z",
      "payload_json": {
        "batteryMv": 4200,
        "temperatureC": 23.5
      }
    }
  ],
  "total": 5000
}
```

## Error Handling

All errors return a JSON response with an error message:

```json
{
  "error": "Description of the error"
}
```

**Common Status Codes:**
- `200 OK` - Request succeeded
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Missing or invalid authentication token
- `500 Internal Server Error` - Server error

## Health Checks

### Health Status
**GET** `/healthz`

Check if the service is running.

**Response:** (200 OK)
```json
{
  "status": "ok"
}
```

### Readiness Status
**GET** `/readyz`

Check if the service is ready to handle requests (MQTT connected).

**Response:** (200 OK or 503 Service Unavailable)
```json
{
  "status": "ready"
}
```

## Usage Examples

### With cURL

**Register:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**List Devices (with token):**
```bash
curl -X GET 'http://localhost:8080/api/v1/devices?limit=10&offset=0' \
  -H "Authorization: Bearer <your_token>"
```

**Get Device Telemetry:**
```bash
curl -X GET 'http://localhost:8080/api/v1/devices/device_001/telemetry?limit=50' \
  -H "Authorization: Bearer <your_token>"
```

### With Flutter

**Register:**
```dart
final response = await http.post(
  Uri.parse('http://localhost:8080/api/v1/auth/register'),
  headers: {'Content-Type': 'application/json'},
  body: jsonEncode({
    'email': 'user@example.com',
    'password': 'password123',
    'name': 'John Doe',
  }),
);

final data = jsonDecode(response.body);
final token = data['token'];
```

**Get Devices:**
```dart
final response = await http.get(
  Uri.parse('http://localhost:8080/api/v1/devices?limit=10'),
  headers: {
    'Authorization': 'Bearer $token',
  },
);

final data = jsonDecode(response.body);
final devices = data['devices'];
```

## Environment Variables

- `JWT_SECRET` - Secret key for JWT signing (default: 'your-super-secret-jwt-key-change-in-production')
- `HTTP_PORT` - HTTP server port (default: 8080)
- `POSTGRES_DSN` - PostgreSQL connection string
- Other MQTT and app configuration variables
