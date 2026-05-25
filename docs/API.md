# Smart Lock API Reference

This document outlines the REST API endpoints exposed by the Go backend (typically listening on `http://localhost:8080`).

---

## 1. System Health & Metrics

### Get Real-Time Service Statuses
Returns the instant connectivity and latency of backing services.

* **URL**: `/api/health`
* **Method**: `GET`
* **Response Body**:
```json
{
  "influxdb": {
    "service": "influxdb",
    "online": true,
    "latency_ms": 2.15
  },
  "mosquitto": {
    "service": "mosquitto",
    "online": true,
    "latency_ms": 0.45
  },
  "postgres": {
    "service": "postgres",
    "online": true,
    "latency_ms": 1.25
  },
  "rabbitmq": {
    "service": "rabbitmq",
    "online": true,
    "latency_ms": 0.85
  }
}
```

### Get Service Health Time-Series Uptime
Queries time-series status pings stored in InfluxDB to render uptime/downtime graphs.

* **URL**: `/api/metrics/health`
* **Method**: `GET`
* **Query Parameters**:
  - `range` (optional, default: `24h`): Duration starts back from now (e.g. `12h`, `24h`, `7d`).
  - `interval` (optional, default: `1m`): Aggregation time window (e.g. `1m`, `5m`, `1h`).
* **Response Body**:
```json
{
  "series": [
    {
      "service": "postgres",
      "points": [
        {"ts": "2026-05-23T10:00:00Z", "status": 1},
        {"ts": "2026-05-23T10:01:00Z", "status": 0}
      ]
    },
    {
      "service": "rabbitmq",
      "points": [
        {"ts": "2026-05-23T10:00:00Z", "status": 1},
        {"ts": "2026-05-23T10:01:00Z", "status": 1}
      ]
    }
  ]
}
```

---

## 2. Telemetry & Control

### Fetch Latest Telemetry for Device
Gets the latest telemetry logs of a specific smart lock device from Postgres.

* **URL**: `/api/telemetry/latest`
* **Method**: `GET`
* **Query Parameters**:
  - `device_id` (required): The unique identifier of the ESP32 hardware device.
* **Response Body**:
```json
{
  "device_id": "lock-1",
  "event": "heartbeat",
  "details": "",
  "status": "online",
  "distance_cm": 12.5,
  "light_level": 40,
  "fails": 0,
  "rssi": -65,
  "uptime": 1200.5
}
```

### Remote Unlock Command
Asynchronously publishes a door opening instruction to the lock MQTT channel.

* **URL**: `/api/door/unlock`
* **Method**: `POST`
* **Response Body**:
```json
{
  "success": true
}
```

---

## 3. Users Management

### List Registered Users & RFID UIDs
Gets the list of registered RFID credentials with optional profile completeness filter.

* **URL**: `/api/users`
* **Method**: `GET`
* **Query Parameters**:
  - `incomplete` (optional): If set to `true`, returns only entries where the user `name` or `email` fields are empty/null.
* **Response Body**:
```json
[
  {
    "id": 1,
    "rfid_uid": "99:88:77:66",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "created_at": "2026-05-24T18:00:00Z",
    "updated_at": "2026-05-24T18:00:00Z"
  },
  {
    "id": 2,
    "rfid_uid": "11:22:33:44",
    "name": null,
    "email": null,
    "created_at": "2026-05-24T18:15:00Z",
    "updated_at": "2026-05-24T18:15:00Z"
  }
]
```

### Update User Details
Binds a human name and email to a registered RFID card uid.

* **URL**: `/api/users/{uid}`
* **Method**: `PUT`
* **Request Body**:
```json
{
  "name": "Jane Smith",
  "email": "jane.smith@example.com"
}
```
* **Response Body**:
```json
{
  "id": 2,
  "rfid_uid": "11:22:33:44",
  "name": "Jane Smith",
  "email": "jane.smith@example.com",
  "created_at": "2026-05-24T18:15:00Z",
  "updated_at": "2026-05-24T18:45:00Z"
}
```

---

## 4. AI Service Commands

### Run AI Model Evaluation
Evaluates the model classification matrices against a static dataset located inside the AI service environment.

* **URL**: `/api/ai/evaluate`
* **Method**: `POST`
* **Request Body**:
```json
{
  "dataset_path": "data/sensor_events.csv"
}
```
* **Response Body**:
```json
{
  "confusion_matrix": [
    [50, 2, 1, 0],
    [3, 40, 2, 0],
    [0, 4, 35, 1],
    [0, 0, 2, 20]
  ],
  "metrics": {
    "accuracy": 0.92,
    "precision_macro": 0.91,
    "recall_macro": 0.90,
    "f1_macro": 0.90
  },
  "binary_metrics": {
    "accuracy": 0.95,
    "precision": 0.94,
    "recall": 0.93,
    "f1": 0.93
  }
}
```

### Run AI Model Retraining
Triggers model training optimization updates over a given number of training epochs.

* **URL**: `/api/ai/retrain`
* **Method**: `POST`
* **Request Body**:
```json
{
  "epochs": 15,
  "dataset_path": "data/sensor_events.csv"
}
```
* **Response Body**:
```json
{
  "success": true,
  "message": "Model retrained successfully over 15 epochs"
}
```
