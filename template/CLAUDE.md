# Kratos Template Project Guide

This is a Go microservice template based on go-kratos framework, following Simple DDD (Domain-Driven Design) architecture.

## Project Structure

```
template/
├── api/                    # Protobuf definitions and generated code
│   ├── protos/             # Proto source files (organized by business domain)
│   │   ├── web/            # Web API protos
│   │   └── event/          # Event/message protos
│   ├── web/                # Generated Go code from web protos
│   └── event/              # Generated Go code from event protos
├── cmd/server/             # Application entry point
│   ├── main.go             # Main function
│   ├── wire.go             # Wire dependency injection definition
│   └── wire_gen.go         # Wire generated code (auto-generated)
├── configs/                # Configuration files (YAML)
├── internal/               # Business code (DDD layers)
│   ├── biz/                # Domain layer (entities, use cases, interfaces)
│   ├── data/               # Data layer (repository implementations)
│   ├── service/            # Application layer (API handlers)
│   ├── server/             # Server setup (HTTP, gRPC)
│   ├── infra/              # Infrastructure layer (DB, Redis, clients)
│   ├── conf/               # Configuration proto and generated code
│   ├── global/             # Global variables and config access
│   ├── crontab/            # Scheduled jobs
│   └── static/             # Static resources
├── third_party/            # Third-party proto dependencies
└── Makefile                # Build commands
```

## DDD Layer Dependencies

**Allowed dependency direction (top to bottom):**
```
api → service → biz → data → infra
```

**Prohibited reverse dependencies:**
- biz CANNOT import service
- service CANNOT import infra directly
- biz CANNOT import data/dao or data/model directly

## Layer Responsibilities

### 1. `api/protos/` - Proto Definitions
- Define HTTP/gRPC interfaces, DTOs (Request/Response), error codes
- NO business logic
- Generate code with: `make api`

### 2. `internal/biz/` - Domain Layer
- Define **entities**, **aggregates**, **use cases**
- Define **interfaces** (not implementations):
  - `IXxxRepo` - repository interfaces for data persistence
  - `IXxxService` - external service interfaces
- Register providers in `biz.go`: `wire.NewSet(NewXxx, ...)`

### 3. `internal/data/` - Data Layer
- Implement `biz.IXxxRepo` interfaces
- Implement `biz.IXxxService` interfaces
- Use infra providers (DB, Redis, HTTP clients)
- Register providers in `data.go`: `wire.NewSet(NewXxxRepo, ...)`

### 4. `internal/infra/` - Infrastructure Layer
- Provide database connections (`PostgresProvider`, `RedisProvider`)
- Provide external clients (`S3Client`, etc.)
- Named as `XxxProvider` or `XxxClient`
- **Never exposed to biz layer**

### 5. `internal/service/` - Application Layer
- API handlers (防腐层 between api and biz)
- Convert DTOs ↔ domain objects
- Call biz use cases
- Generate service stubs: `kratos proto server api/protos/web/xxx.proto -t internal/service`

## Common Commands

```bash
# Initialize development environment
make init

# Generate proto code (api + config)
make all

# Generate API proto only
make api

# Generate config proto
make config

# Generate Wire and go:generate
make generate

# Generate DAO code
make dao

# Build binary
make build
```

## Wire Dependency Injection

All layers use Google Wire for DI. Each layer has a `ProviderSet`:

```go
// biz/biz.go
var ProviderSet = wire.NewSet(NewAuth, NewProbe)

// data/data.go
var ProviderSet = wire.NewSet(NewAuthRepo, NewHealthRepo)

// service/service.go
var ProviderSet = wire.NewSet(NewAuthService, NewProbeService)

// infra/infra.go
var ProviderSet = wire.NewSet(NewDataProvider, NewS3Client)
```

After adding new providers, run: `make generate`

## Code Conventions

### Naming
- Biz interfaces: `IXxxRepo`, `IXxxService`
- Data implementations: `XxxRepo`, `Xxx` (implements biz interface)
- Infra providers: `XxxProvider`, `XxxClient`
- Service handlers: `XxxService`

### File Organization
- Each biz entity/aggregate in its own file
- Each data repo implementation in its own file
- Mock files in `mocks/` subdirectory
- Tests in `tests/` subdirectory

### Interface Definition Pattern
```go
// In biz/dep.go or biz/xxx.go
type IXxxRepo interface {
    Method(ctx context.Context, args...) (result, error)
}

//go:generate mockgen -source=xxx.go -destination=./mocks/xxx.go -package=mocks
type IXxxService interface {
    Method(ctx context.Context, args...) (result, error)
}
```

### Repository Implementation Pattern
```go
// In data/xxx_repo.go
type XxxRepo struct {
    dbProvider  infra.PostgresProvider
    rdbProvider infra.RedisProvider
    log         *log.Helper
}

func NewXxxRepo(db infra.PostgresProvider, rdb infra.RedisProvider, logger log.Logger) biz.IXxxRepo {
    return &XxxRepo{
        dbProvider:  db,
        rdbProvider: rdb,
        log:         log.NewHelper(logger),
    }
}
```

## Configuration

- Config proto: `internal/conf/conf.proto`
- Config YAML: `configs/config.yaml`
- Environment variables for secrets (e.g., `PG_DSN`)

## Testing

- Unit tests in `xxx_test.go` files
- Integration tests in `tests/` subdirectory
- Use mockgen for mocking interfaces
- Generate mocks: `go generate ./...`

## Adding New Features

1. **Define proto** in `api/protos/web/xxx.proto`
2. **Run** `make api` to generate code
3. **Generate service stub**: `kratos proto server api/protos/web/xxx.proto -t internal/service`
4. **Define biz interface** in `internal/biz/`
5. **Implement data layer** in `internal/data/`
6. **Add to ProviderSets** in each layer's `xxx.go`
7. **Run** `make generate` to update Wire

## Health Checks

- Liveness: `GET /health/live` - returns 200 if process is alive
- Readiness: `GET /health/ready` - returns 200 if all dependencies are healthy
- Detailed: `GET /health` - returns JSON with component status

## Dependencies

Key external dependencies:
- `github.com/go-kratos/kratos/v2` - Kratos framework
- `github.com/google/wire` - Dependency injection
- `gorm.io/gorm` - ORM
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/seanbit/kratos/webkit` - Common utilities (logging, metrics, interceptor)
