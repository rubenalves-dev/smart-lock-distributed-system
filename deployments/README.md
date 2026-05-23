# 🐳 Infrastructure & Deployment

Configuration files for the distributed environment.

### 📦 Services included:

- **PostgreSQL**: Stores logs and user credentials.
- **RabbitMQ**: Handles asynchronous event streaming.
- **Main Server & AI Service**: Containerized application logic.

### 🖥️ Local UI

- **RabbitMQ Management**: [http://localhost:15672](http://localhost:15672) (guest/guest)
- **Main API**: [http://localhost:8080](http://localhost:8080)

### 🔐 InfluxDB auth note

If you see `authorization not found` or 401 write errors, the InfluxDB volumes were initialized with a different token. Recreate the volumes so the init values in `docker-compose.yml` are applied:

```bash
docker-compose -f ./deployments/docker-compose.yml -p smartlock down -v
```

Then bring the stack up again with `make up`.

### 🔁 InfluxDB token reset (without wiping volumes)

If the token still fails (401 on write), create a new all-access token in the running InfluxDB and update the app to use it:

```bash
# 1) Create a new token
NEW_TOKEN=$(docker exec -it influx_db influx auth create --org "A MinhaOrganizacao" --all-access | awk -F'token:' '/token:/ {print $2}' | xargs)

# 2) Update docker-compose.yml (replace the token value)
# INFLUXDB_TOKEN: "<NEW_TOKEN>"

# 3) Recreate the main server container
/docker-compose -f ./deployments/docker-compose.yml -p smartlock up -d --force-recreate main-server

# 4) Verify writes work
/docker exec -it influx_db influx write \
  --org "A MinhaOrganizacao" \
  --bucket "esp32_pings" \
  --token "<NEW_TOKEN>" \
  "service_health,service=test status=1,latency_ms=0.1"
```

If step 4 still returns 401, double‑check the org and bucket names using:

```bash
docker exec -it influx_db influx org list
docker exec -it influx_db influx bucket list
```
