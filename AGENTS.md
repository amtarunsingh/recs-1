# Recs Votes Storage - Agent Reference

Go 1.25 | Clean Architecture + DDD | REST API + Event-Driven Worker

## Quick Reference

```bash
make dev-up          # Start all services (localstack, infra, API, worker)
make test            # Run tests
make test-coverage   # 95% minimum coverage
make wire            # Regenerate DI after changes
```

API: http://localhost:8888 | Docs: http://localhost:8888/docs

## Architecture

```
internal/context/voting/
├── domain/              # Entities, VOs, repos (interfaces)
├── application/         # Operations (use cases), services
├── infrastructure/      # DynamoDB repos, AWS clients
└── interface/api/rest/  # HTTP contracts, DTOs
```

**Dependency Rule:** Domain ← Application ← Infrastructure/Interface

## Domain Model

**Romance** (aggregate root)
- `ActiveUserVote`, `PeerUserVote` (Vote entities)
- `Version` (optimistic locking)
- PK: user_id, SK: peer_id

**Vote** (entity)
- `VoteType`: empty(0), yes(1), no(2), crush(3), compliment(4)
- Categories: Positive(1,3,4), Negative(2), Empty(0)

**CountersGroup** (VO)
- `IncomingYes`, `IncomingNo`, `OutgoingYes`, `OutgoingNo`
- Hour-based + lifetime aggregates

## API Endpoints

Base: `/v1`

**Votes:**
- `GET /votes/{country_id}/{active_user_id}/{peer_id}` - Get vote
- `POST /votes/{country_id}` - Add vote
- `PATCH /votes/{country_id}/{active_user_id}/{peer_id}/change-contract` - Change vote
- `DELETE /votes/{country_id}/{active_user_id}/{peer_id}` - Delete vote

**Romances:**
- `GET /romances/{country_id}/{active_user_id}/{peer_id}` - Get romance
- `DELETE /romances/{country_id}/{active_user_id}/{peer_id}` - Delete romance
- `DELETE /romances/{country_id}/{active_user_id}` - Delete all (async)

**Counters:**
- `GET /counters/{country_id}/{active_user_id}/lifetime` - Lifetime counters
- `GET /counters/{country_id}/{active_user_id}/hourly?hours_offset_groups=[...]` - Hourly counters

## DynamoDB Schema

**Romances Table:**
- PK: `a` (user_id), SK: `b` (peer_id)
- GSI: `gsiByMaxMinUser` (swapped PK/SK)
- Attrs: `e,g,h,i` (PK user vote), `l,n,o,p` (SK user vote), `v` (version), `ttl`
- TTL: Mutual(546d), Non-mutual(180d), Dead(90d)
- Repo: `internal/context/voting/infrastructure/persistence/romances_repository.go:619`

**Counters Table:**
- PK: `u` (user_id), SK: `h` (hour_timestamp, 0=lifetime)
- Attrs: `iy,in,oy,on` (incoming/outgoing yes/no), `ttl` (48h for hourly)
- Repo: `internal/context/voting/infrastructure/persistence/counters_repository.go:305`

## Key Operations

Location: `internal/context/voting/application/operation/`

- `AddUserVoteOperation` - Validates transitions, retries on version conflict (3x), updates counters
- `ChangeUserVoteOperation` - Changes vote type, adjusts counters
- `DeleteRomancesOperation` - Streams peer IDs via channel, batches 25 per message
- `DeleteRomancesGroupOperation` - Batch delete via DynamoDB BatchWriteItem

Service: `internal/context/voting/application/voting_service.go:181`

## Event-Driven Flow

**Delete all romances:**
1. `DELETE /romances/{country}/{user}` → publish to `delete-romances.fifo`
2. Worker receives → `DeleteRomancesOperation` streams peer IDs
3. Batch 25 IDs → publish to `delete-romances-group.fifo`
4. `DeleteRomancesGroupHandler` → DynamoDB BatchWriteItem

**Topics:** `delete-romances.fifo`, `delete-romances-group.fifo`
**Queues:** `delete-romances-queue.fifo`, `delete-romances-group-queue.fifo`
**Library:** Watermill

## Configuration

Env vars:
- `LOG_LEVEL=INFO`
- `AWS_REGION=us-east-2`
- `DYNAMO_DB_ENDPOINT=http://localstack:4566` (local)
- `SNS_ENDPOINT=http://localstack:4566` (local)

Constants:
- `CountersTtlHours: 48`
- `DynamoDbVersionConflictRetriesCount: 3`
- TTL: Mutual(546d), Non-mutual(180d), Dead(90d)

## Testing

**Integration tests:** TestContainers + LocalStack 4.9.0
- Location: `internal/test/integration/`
- Setup: `internal/testlib/testcontainer/localstack.go`
- Helpers: `internal/testlib/helper/`

**Mocks:** `internal/testlib/mocks/` (uber-go/mock)

**Coverage target:** 95%

## Coding Conventions

**Constructors:**
- Return pointers for services/repos (pointer receivers)
- Return values for small immutable VOs

**Error wrapping:**
- Wrap at subsystem boundaries (API→Service→Repo)
- Skip if error already clear

**File order:**
1. Package declaration
2. Imports (stdlib, external, internal)
3. Constants
4. Type definitions
5. Constructors
6. Methods (grouped by receiver)
7. Private helpers

**Naming:**
- Interfaces: `Repository`, `Publisher`, `Logger`
- Implementations: `DynamoDBRepository`, `SnsPublisher`
- Operations: `AddUserVoteOperation`
- Handlers: `DeleteRomancesHandler`
- Messages: `DeleteRomancesMessage`

**DI:** Google Wire (`make wire` after changes)

## Tech Stack

| Component | Tech | Version |
|-----------|------|---------|
| REST | Huma | v2.34.1 |
| DB | DynamoDB | aws-sdk-go-v2 v1.52.1 |
| Messaging | Watermill + SNS/SQS | v1.5.1 / v1.0.1 |
| Infra | AWS CDK (Go) | v2.220.0 |
| DI | Wire | v0.7.0 |
| Testing | TestContainers + LocalStack | v0.39.0 |
| Mocking | uber-go/mock | v0.6.0 |

## Common Tasks

**Add new operation:**
1. Create in `internal/context/voting/application/operation/`
2. Add method to `VotingService`
3. Create API handler in `internal/context/voting/interface/api/rest/v1/`
4. Register route in `register.go`
5. Run `make wire` if DI changes

**Add new message handler:**
1. Create message in `application/messaging/message/`
2. Create handler in `application/messaging/handler/`
3. Register in `internal/app/bootstrap/handlers.go`
4. Add topic constant in operations
5. Run `make wire`

**Run single test:**
```bash
go test ./internal/test/integration/voting/persistence/ -run TestName -v
```

## File Locations

- API routes: `internal/context/voting/interface/api/rest/v1/register.go`
- DI wire: `internal/app/di/wire.go`
- Message bootstrap: `internal/app/bootstrap/handlers.go`
- Message processor: `internal/app/message_processor.go`
- Server: `internal/app/server.go`
- Config: `config/config.go:60`
- CDK stack: `infra/data_stack.go:99`