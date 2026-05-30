# Linear Relay — SuperPlane Webhook Receiver

A standalone Go webhook relay service that receives Linear webhook events, validates their HMAC-SHA256 signatures, deduplicates them via PostgreSQL, transforms the payload, and forwards it to SuperPlane.

## How It Works

1. Linear sends a webhook payload to `POST /webhook`
2. The relay validates the `X-Linear-Signature` header using HMAC-SHA256 with the `LINEAR_WEBHOOK_SECRET`
3. The event ID is checked against PostgreSQL for idempotency (deduplication)
4. The payload is transformed into SuperPlane format
5. The transformed payload is forwarded to `SUPERPLANE_WEBHOOK_URL` via HTTP POST
6. A `200 OK` is returned to Linear immediately (Linear retries on non-2xx)

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `PORT` | No | HTTP listen port (default: `8080`). Render sets this automatically. |
| `DATABASE_URL` | Yes | PostgreSQL connection string (provided by Render PostgreSQL). |
| `LINEAR_WEBHOOK_SECRET` | Yes | Secret used to verify HMAC-SHA256 signatures from Linear webhooks. |
| `SUPERPLANE_WEBHOOK_URL` | Yes | SuperPlane endpoint URL to forward transformed events to. |

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check (verifies database connectivity). Returns `200 ok`. |
| `POST` | `/webhook` | Receives Linear webhook payloads. |

## Database Schema

On startup, the relay creates the following table if it does not exist:

```sql
CREATE TABLE IF NOT EXISTS webhook_events (
    event_id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Payload Transformation

The relay transforms Linear webhook payloads into the following format before forwarding:

```json
{
  "source": "linear",
  "type": "<Linear event type>",
  "action": "<Linear action>",
  "data": { "<original Linear data>" },
  "timestamp": "<Linear createdAt>",
  "event_id": "<Linear event id, if present>"
}
```

## Deploy to Render

### Prerequisites

1. A [Render](https://render.com) account
2. A SuperPlane instance with a webhook URL

### Steps

1. **Create a Render PostgreSQL database**
   - In the Render Dashboard, create a new PostgreSQL instance
   - Note the `Internal Database URL` — this becomes `DATABASE_URL`

2. **Create a Render Web Service**
   - Connect your Git repository containing this service
   - Set **Build Command**: leave blank (Dockerfile handles the build)
   - Set **Start Command**: leave blank (Dockerfile CMD handles it)
   - Choose **Docker** as the environment

3. **Set environment variables** in the Render Web Service:
   - `DATABASE_URL` — Internal Database URL from step 1
   - `LINEAR_WEBHOOK_SECRET` — from your Linear webhook integration settings
   - `SUPERPLANE_WEBHOOK_URL` — your SuperPlane webhook ingest URL

4. **Configure the Linear webhook**
   - In Linear, create a webhook pointing to your Render service URL: `https://<your-service>.onrender.com/webhook`
   - Note the HMAC signature secret and set it as `LINEAR_WEBHOOK_SECRET`

### Render Blueprint (Optional)

You can also use a `render.yaml` blueprint:

```yaml
services:
  - type: web
    name: linear-relay
    env: docker
    plan: starter
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: linear-relay-db
          property: connectionString
      - key: LINEAR_WEBHOOK_SECRET
        sync: false
      - key: SUPERPLANE_WEBHOOK_URL
        sync: false

databases:
  - name: linear-relay-db
    plan: starter
```

## Running Locally

### With Docker

```bash
docker build -t linear-relay .
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@localhost:5432/relay?sslmode=disable" \
  -e LINEAR_WEBHOOK_SECRET="your-secret" \
  -e SUPERPLANE_WEBHOOK_URL="https://superplane.example.com/webhook" \
  linear-relay
```

### Without Docker

```bash
# Ensure a PostgreSQL instance is running and accessible
export DATABASE_URL="postgres://user:pass@localhost:5432/relay?sslmode=disable"
export LINEAR_WEBHOOK_SECRET="your-secret"
export SUPERPLANE_WEBHOOK_URL="https://superplane.example.com/webhook"

go run .
```

## Testing Locally

Health check:

```bash
curl http://localhost:8080/health
# Expected: ok
```

Send a test webhook:

```bash
# Compute the HMAC signature for your test payload
PAYLOAD='{"type":"Issue","action":"create","createdAt":"2024-01-01T00:00:00Z","data":{"id":"abc123","title":"Test issue"}}'

SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "your-secret" | awk '{print $2}')

curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-Linear-Signature: $SIGNATURE" \
  -d "$PAYLOAD"
# Expected: ok
```

Send a duplicate event (should be deduplicated):

```bash
# Same request again
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-Linear-Signature: $SIGNATURE" \
  -d "$PAYLOAD"
# Expected: duplicate
```

## Notes

- The relay always returns `200 OK` to Linear, even if forwarding to SuperPlane fails. This prevents Linear from retrying and creating duplicates. Events that fail to forward are still recorded as seen.
- Old deduplication records should be periodically cleaned up. You can add a cron job or scheduled query to delete rows older than a threshold:

```sql
DELETE FROM webhook_events WHERE created_at < NOW() - INTERVAL '7 days';
```

## License

Part of the SuperPlane project.