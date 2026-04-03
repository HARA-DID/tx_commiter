# worker-service

A production-ready Go worker that consumes events from **Redis Streams**, stores
job state in **PostgreSQL**, and executes blockchain transactions via a **Composite SDK Adapter** (DID Root, Verifiable Credentials, Alias) with **Account Abstraction (AA)** routing.

```
Redis Stream ──XREADGROUP──► Worker Pool ──► EventService ──► CompositeAdapter 
                                   │                              │      │
                                   │                              │      └─► SDK Adapters (Encode)
                                   ▼                              │              │
                             PostgreSQL (jobs)                    └─► AAAdapter (HandleOps)
                                                                         │
                                                                         ▼
                                                              Blockchain (Hara Chain)
```

## Folder Structure & File Responsibilities

```
.
├── cmd/
│   ├── worker/              # Main application entrypoint
│   └── dlq-reader/          # Ops tool: tracks and reads the dead-letter queue (DLQ)
│
├── internal/
│   ├── config/              # Env-based configuration
│   │   └── config.go        # Loads and validates HNS-based variables
│   │
│   ├── domain/              # Core business types and payloads (Shared)
│   │   ├── did.go           # DID Root registry payloads
│   │   ├── vc.go            # Verifiable Credentials payloads
│   │   ├── alias.go         # Alias registration payloads
│   │   ├── aa.go            # Account Abstraction / HandleOps payloads
│   │   └── job.go           # Job state and status definitions
│   │
│   ├── sdk/                 # Blockchain Integration Layer (The only place SDKs are imported)
│   │   ├── composite_adapter.go # Routes jobs to the correct SDK adapter
│   │   ├── did_adapter.go   # DID Root SDK implementation (HNS-only)
│   │   ├── vc_adapter.go    # VC SDK implementation (HNS-only)
│   │   ├── alias_adapter.go # Alias SDK implementation (HNS-only)
│   │   ├── aa_adapter.go    # AA EntryPoint implementation (HandleOps)
│   │   └── provider.go      # Shared blockchain client/wallet setup
│   │
│   ├── service/             # Orchestration Layer
│   │   ├── event_service.go # Main logic: idempotency -> database -> SDK routing
│   │   └── blockchain.go    # Generic interface for all blockchain operations
│   │
│   ├── infra/
│   │   ├── db/              # PostgreSQL + JobRepository (persistence)
│   │   └── redis/           # Redis Stream consumer & DLQ management
│   │
│   ├── worker/              # Consumer loop, error handling, and Prometheus metrics
│   └── mocks/               # Mock implementations for testing
│
├── pkg/                     # Shared utilities (Retry, Metrics)
├── Dockerfile
├── docker-compose.yml
└── .env.example             # Template for all required HNS & infrastructure variables
```

---

## HNS Contract Resolution

This project exclusively uses **Handshake (HNS)** for contract resolution. There are no hardcoded addresses or manual ABI configurations.
All adapters use `NewXXXWithHNS` or `ContractWithHNS` to resolve dependencies at startup via the `AA_ENTRYPOINT_HNS`, `DID_VC_FACTORY_HNS`, and other HNS environment variables.

## Environment Variables

| Variable            | Required | Description                                           |
|---------------------|----------|-------------------------------------------------------|
| `REDIS_URL`         | ✅       | Redis connection URL                                  |
| `STREAM_NAME`       | ✅       | Redis stream to consume                               |
| `GROUP_NAME`        | ✅       | Consumer group name                                   |
| `DB_URL`            | ✅       | PostgreSQL DSN                                        |
| `RPC_URLS`          | ✅       | Comma-separated blockchain RPC endpoints              |
| `PRIVATE_KEY`       | ✅       | Hex-encoded wallet private key                        |
| `AA_ENTRYPOINT_HNS` | ✅       | HNS URI for the Accountant Abstraction EntryPoint     |
| `DID_VC_FACTORY_HNS`| ✅       | HNS URI for the Verifiable Credentials Factory        |
| `DID_ALIAS_FACTORY_HNS`| ✅    | HNS URI for the Alias Factory                         |
| `DID_ROOT_FACTORY_HNS`| ✅     | HNS URI for the DID Root Factory                      |
| `WORKER_CONCURRENCY`| ❌       | Max concurrent goroutines (Default: 10)               |
| `MAX_RETRY`         | ❌       | Max blockchain retry attempts per event (Default: 3) |
| `SERVER_PORT`       | ❌       | Port for health and metrics (Default: 8080)           |

---

## Processing Flow

For each Redis stream message the worker:

1. **Parses** the raw stream entry into `domain.Event`.
2. **Validates** the event (id present, type recognised, payload non-empty).
3. **Idempotency check** — queries `jobs` table by `event_id`; skips if already `success`.
4. **Creates** a `pending` job row in PostgreSQL.
5. **Encodes** the transaction data by mapping domain payloads to SDK-specific `Params`.
6. **Dispatches** via the **AA EntryPoint** (`HandleOps`) with retry + exponential backoff.
7. **Updates** the job row to `success` (with tx hashes) or `failed` (with error).
8. **ACKs** the message if successful; pushes to DLQ and ACKs if retries are exhausted.

---

## Observability

### Health check

```
GET http://localhost:8080/healthz
→ 200 {"status":"ok"}
```

### Prometheus metrics

```
GET http://localhost:8080/metrics
```

Key metrics:

| Metric                                  | Type      | Description                              |
|-----------------------------------------|-----------|------------------------------------------|
| `worker_events_received_total`          | Counter   | Events read from Redis                   |
| `worker_events_processed_total{status}` | Counter   | Events by outcome: success/failed/skipped|
| `worker_events_retried_total`           | Counter   | Total retry attempts                     |
| `worker_events_dlq_total`               | Counter   | Events routed to DLQ                     |
| `worker_event_process_duration_seconds` | Histogram | End-to-end processing latency            |

---

## Horizontal Scaling

Each replica **must** have a unique `CONSUMER_NAME`. With the default
`worker-<hostname>` value, Kubernetes pods get unique names automatically.

All replicas join the **same consumer group** — Redis Streams guarantees
each message is delivered to exactly one consumer in the group.

```yaml
# k8s Deployment excerpt
env:
  - name: CONSUMER_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name   # e.g. worker-service-7d9f8b-xkqp2
```

---

## Running Tests

```bash
make test
make cover
```

## Integration Guide

Events are sent to the worker via **Redis Streams**. Each stream entry must contain a `payload` field with a JSON string representing the event.

### Message Format

```bash
# XADD <stream_name> * payload <json_event>
XADD hara.events * payload '{"id": "uuid-123", "type": "CREATE_DID", "payload": {"did": "did:hara:123", "target_address": "0x..."}}'
```

### Event Payload Examples

#### DID: Create New DID
```json
{
  "id": "evt_01",
  "type": "CREATE_DID",
  "payload": {
    "did": "did:hara:user_01",
    "target_address": "0x123...",
    "multiple_rpc_calls": false
  }
}
```

#### VC: Issue Credential
```json
{
  "id": "evt_02",
  "type": "ISSUE_CREDENTIAL",
  "payload": {
    "option": 1,
    "did_recipient": "did:hara:user_01",
    "issuer": "0xIssuer...",
    "expired_at": 1735689600,
    "offchain_hash": "0xHash...",
    "target_address": "0x123..."
  }
}
```

#### Alias: Register Domain
```json
{
  "id": "evt_03",
  "type": "REGISTER_DOMAIN",
  "payload": {
    "node": "0xNodeHash...",
    "label": "my-domain",
    "duration": 31536000,
    "target_address": "0x123..."
  }
}
```

#### Org: Add Member
```json
{
  "id": "evt_04",
  "type": "ADD_MEMBER",
  "payload": {
    "org_did_index": 5,
    "data": "0xEncodedMemberData...",
    "target_address": "0x123..."
  }
}
```

---

## Architecture Decisions

**SDK Isolation**: The `internal/sdk/` directory is the **only** entry point for third-party SDK dependencies (DID, VC, Alias, AA). The rest of the application interacts with the blockchain via a high-level `BlockchainService` interface.

**Composite SDK Pattern**: A `CompositeAdapter` acts as a router, delegating jobs to specific SDK adapters based on the event type.

**Account Abstraction (AA) Integration**: All write operations follow an **Encode-then-Dispatch** pattern. Specific adapters (DID, VC, Alias) encode their parameters into binary calldata, which is then passed to the `AAAdapter`.

**HNS-Only Configuration**: This project exclusively uses **Handshake (HNS)** for contract resolution. There are no hardcoded addresses or manual ABI configurations.

**Event Idempotency**: To prevent double-processing of events, we combine an application-level check with a database-level `unique_event_id` constraint.

**ACK-after-Success/DLQ Policy**: Messages are only acknowledged in Redis after a successful blockchain transaction or after being successfully pushed to the Dead Letter Queue (DLQ).
