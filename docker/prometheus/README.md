# Prometheus Monitoring Setup

This directory contains the Prometheus configuration for monitoring the votes-storage service.

## Quick Start

### Start all services including Prometheus
```bash
make dev-up
```

### Start only Prometheus (if app is already running)
```bash
make prometheus-up
```

### Stop Prometheus
```bash
make prometheus-down
```

### Reload Prometheus configuration
After modifying `prometheus.yml`:
```bash
make prometheus-reload
```

## Access

- **Prometheus UI**: http://localhost:9090
- **Metrics endpoint**: http://localhost:8888/metrics

## Available Metrics

### HTTP Request Metrics
- `http_request_duration_seconds` - Histogram tracking request latency with buckets
- `http_requests_total` - Counter for total HTTP requests by method, path, and status

### Vote Operation Metrics
- `votes_total{vote_type, operation}` - Total votes processed by type and operation
  - vote_type: yes, no, crush, compliment, empty
  - operation: add, change, delete

- `vote_errors_total{operation, error_type}` - Total vote operation errors
  - operation: add, change, delete
  - error_type: duplicate, invalid_transition, get_romance_error, db_error

- `vote_changes_total{from_type, to_type}` - Vote type transitions
  - Tracks changes like yes→crush, no→yes, etc.

- `vote_deletions_total{vote_type}` - Total vote deletions by type

## Example Queries

### Request rate over last 5 minutes
```promql
rate(http_requests_total[5m])
```

### P95 request latency
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

### Total votes by type
```promql
sum by (vote_type) (votes_total)
```

### Error rate by operation
```promql
rate(vote_errors_total[5m])
```

### Vote changes from yes to crush
```promql
vote_changes_total{from_type="yes", to_type="crush"}
```

## Configuration

The Prometheus configuration is located in `prometheus.yml`. It scrapes the votes-storage API service every 15 seconds.

Key settings:
- Scrape interval: 15s
- Target: app:8888/metrics
- Job name: votes-storage-api

## Data Persistence

Prometheus data is stored in a Docker volume named `prometheus-data`. To reset all metrics data:

```bash
docker compose -f docker/docker-compose.dev.yml down -v
```

**Warning**: This will delete all Prometheus historical data.
