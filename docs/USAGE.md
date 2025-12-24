# Codo Framework Usage Guide

This comprehensive reference enables Claude and developers to quickly understand and build applications with the Codo Framework.

## Table of Contents

1. [Quick Start: Creating a New App](#1-quick-start-creating-a-new-app)
2. [Framework Features](#2-framework-features)
3. [Middleware Creation & Registration](#3-middleware-creation--registration)
4. [Client Creation & Registration](#4-client-creation--registration)
5. [Environment Variables & Configuration](#5-environment-variables--configuration)
6. [Tasks](#6-tasks)
7. [Models, Migrations, Seeds](#7-models-migrations-seeds)
8. [Forms, Services, Handlers](#8-forms-services-handlers)
9. [Logging](#9-logging)
10. [Advanced Features](#10-advanced-features)

---

## 1. Quick Start: Creating a New App

### Project Structure

```
myapp/
├── main.go                          # Entry point
├── config.yaml                      # Configuration file
├── .env                             # Environment variables (optional)
├── internal/
│   ├── app/
│   │   ├── app.go                   # Bootstrap options
│   │   └── features.go              # Feature flags (optional)
│   └── clients/
│       └── clients.go               # Custom client registration
└── pkg/
    ├── handlers/
    │   ├── register.go              # Handler registration
    │   └── user_handler.go          # Example handler
    ├── services/
    │   └── user_service.go          # Business logic
    ├── forms/
    │   └── user_forms.go            # DTOs
    ├── models/
    │   └── user.go                  # Domain models
    ├── migrations/
    │   └── migrations.go            # Database migrations
    └── seeds/
        └── seeds.go                 # Database seeds
```

### main.go

```go
package main

import (
    "fmt"
    "os"

    "github.com/codoworks/codo-framework/cmd"
    _ "github.com/codoworks/codo-framework/cmd/auto"  // Auto-register CLI commands
    "github.com/codoworks/codo-framework/core/app"

    internalApp "myapp/internal/app"
)

// Version is injected at build time via -ldflags
var Version = "0.0.1-dev"

func init() {
    // Set version in app package
    internalApp.Version = Version

    // Register app initializer with framework
    app.RegisterInitializer(internalApp.Initialize)

    // Register CLI metadata
    meta := internalApp.AppMetadata()
    cmd.SetAppInfo(meta.Name(), meta.Short(), meta.Long())
    cmd.SetVersion(meta.Version())
}

func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### internal/app/app.go

```go
package app

import (
    "github.com/codoworks/codo-framework/core/app"
    "github.com/codoworks/codo-framework/core/config"
    "github.com/codoworks/codo-framework/core/metadata"

    internalClients "myapp/internal/clients"
    "myapp/pkg/handlers"
    "myapp/pkg/migrations"
)

// Version is set by main.go
var Version = "0.0.0-dev"

// Initialize returns BootstrapOptions for the framework.
// The framework CLI commands set the Mode field based on the command.
func Initialize(cfg *config.Config) (app.BootstrapOptions, error) {
    return app.BootstrapOptions{
        HandlerRegistrar: handlers.RegisterAll,           // Register HTTP handlers
        MigrationAdder:   migrations.AddToRunner,         // Add database migrations
        CustomClientInit: internalClients.Initialize,     // Initialize custom clients
        EnvVarRegistrar:  internalClients.RegisterEnvVars, // Register env vars
    }, nil
}

// AppMetadata returns CLI metadata without app initialization.
func AppMetadata() metadata.Metadata {
    return metadata.Info{
        AppName:    "myapp",
        AppShort:   "My Application CLI",
        AppLong:    "My Application - A service built with Codo Framework",
        AppVersion: Version,
    }
}
```

### internal/clients/clients.go

```go
package clients

import (
    "fmt"

    "github.com/codoworks/codo-framework/clients/logger"
    "github.com/codoworks/codo-framework/core/clients"
    "github.com/codoworks/codo-framework/core/config"
)

// RegisterMetadata registers metadata for custom clients.
func RegisterMetadata() {
    // Example: Mark a client as required
    clients.RegisterMetadata(clients.ClientMetadata{
        Name:        "my-client",
        Requirement: clients.ClientRequired,
    })
}

// RegisterEnvVars registers environment variable requirements.
// Called during bootstrap BEFORE client initialization.
func RegisterEnvVars(registry *config.EnvVarRegistry) error {
    return registry.RegisterMany([]config.EnvVarDescriptor{
        {
            Name:        "MY_API_KEY",
            Type:        config.EnvTypeString,
            Required:    true,
            Group:       "myservice",
            Description: "API key for my service",
            Sensitive:   true,  // Masked in logs
        },
        {
            Name:        "MY_API_URL",
            Type:        config.EnvTypeURL,
            Required:    true,
            Group:       "myservice",
            Description: "Base URL for my service",
        },
    })
}

// Initialize registers and initializes all custom clients.
func Initialize(cfg *config.Config) error {
    RegisterMetadata()

    log := getOrCreateLogger()

    // Get resolved env vars from registry
    myVars := cfg.EnvRegistry.GetGroup("myservice")

    // Example: Initialize a custom client
    // myClient := myclient.New()
    // clients.MustRegister(myClient)
    // myClient.Initialize(&myclient.Config{
    //     APIKey: myVars["MY_API_KEY"].String(),
    //     URL:    myVars["MY_API_URL"].URL().String(),
    // })

    log.Info("Custom clients initialized")
    return nil
}

func getOrCreateLogger() *logger.Logger {
    if clients.Has(logger.ClientName) {
        log, err := clients.GetTyped[*logger.Logger](logger.ClientName)
        if err == nil {
            return log
        }
    }
    log := logger.New()
    _ = log.Initialize(nil)
    return log
}
```

### pkg/handlers/register.go

```go
package handlers

import (
    "fmt"

    "github.com/codoworks/codo-framework/clients/logger"
    "github.com/codoworks/codo-framework/core/clients"
    "github.com/codoworks/codo-framework/core/db"
    "github.com/codoworks/codo-framework/core/http"

    "myapp/pkg/services"
)

// RegisterAll registers all handlers with the HTTP framework.
func RegisterAll() error {
    log, _ := clients.GetTyped[*logger.Logger](logger.ClientName)

    // Get database client
    dbClient, err := clients.GetTyped[*db.Client]("db")
    if err != nil {
        return fmt.Errorf("database client required: %w", err)
    }

    // Create services
    userService := services.NewUserService(dbClient)

    // Register handlers
    http.RegisterHandler(NewUserHandler(userService))

    if log != nil {
        log.Info("All handlers registered successfully")
    }

    return nil
}
```

### CLI Commands

```bash
# Build
go build -ldflags "-X main.Version=1.0.0" -o myapp .

# Start all servers (public:8081, protected:8080, hidden:8079)
./myapp serve

# Start individual servers
./myapp start public
./myapp start protected
./myapp start hidden

# Database operations
./myapp db migrate
./myapp db rollback
./myapp db seed

# Information
./myapp info routes
./myapp info env

# Tasks
./myapp task list
./myapp task run user:promote-admin --user-id=123
```

---

## 2. Framework Features

### 2.1 Validation System

**Source:** `core/forms/validation.go`

Uses `go-playground/validator` with automatic error transformation.

```go
import "github.com/codoworks/codo-framework/core/forms"

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email,max=255"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8,max=72"`
    Role     string `json:"role" validate:"omitempty,oneof=user admin moderator"`
    Age      int    `json:"age" validate:"omitempty,gte=0,lte=150"`
}

// Validate in handler
func (h *UserHandler) Create(c *http.Context) error {
    var form CreateUserRequest
    if err := c.BindAndValidate(&form); err != nil {
        return err  // Automatic structured error response
    }
    // ... use validated form
}

// Manual validation
err := forms.Validate(&form)
if err != nil {
    validationErrs := err.(*forms.ValidationErrors)
    for _, fieldErr := range validationErrs.Errors {
        fmt.Printf("Field: %s, Message: %s\n", fieldErr.Field, fieldErr.Message)
    }
}

// Register custom validator
forms.RegisterValidation("my_custom", func(fl validator.FieldLevel) bool {
    return fl.Field().String() == "expected"
})
```

**Available Validators:**
- `required` - Field must be present
- `email` - Valid email format
- `url` - Valid URL
- `uuid` - Valid UUID
- `min=N` / `max=N` - String length or numeric range
- `len=N` - Exact length
- `gte=N` / `lte=N` / `gt=N` / `lt=N` - Numeric comparisons
- `oneof=a b c` - Value must be one of listed options
- `alpha` / `alphanum` / `numeric` - Character constraints
- `eqfield=OtherField` - Must equal another field

### 2.2 Error Handling

**Source:** `core/errors/errors.go`, `core/errors/codes.go`

Hierarchical error codes with HTTP status mapping:

```go
import "github.com/codoworks/codo-framework/core/errors"

// Creating errors
err := errors.NotFound("User not found")
err := errors.BadRequest("Invalid user ID format")
err := errors.Unauthorized("Invalid session")
err := errors.Forbidden("Access denied")
err := errors.Conflict("Email already in use")
err := errors.Internal("Database connection failed")
err := errors.Validation("Invalid input", []string{"email", "name"})
err := errors.Timeout("Request timed out")
err := errors.Unavailable("Service temporarily unavailable")

// Fluent error building
err := errors.NotFound("User not found").
    WithCause(dbErr).
    WithDetail("user_id", userID).
    WithField("id", "VALIDATION.UUID_INVALID", "Must be a valid UUID").
    WithUserMessage("We couldn't find that user")

// Wrapping errors
err := errors.WrapInternal(dbErr, "Failed to query database")
err := errors.WrapNotFound(dbErr, "User lookup failed")
err := errors.WrapBadRequest(parseErr, "Invalid request body")

// Checking error types
if errors.IsNotFound(err) { ... }
if errors.IsUnauthorized(err) { ... }
if errors.IsBadRequest(err) { ... }
if errors.IsConflict(err) { ... }
if errors.IsValidation(err) { ... }

// Get HTTP status
status := errors.GetHTTPStatus(err)  // e.g., 404
code := errors.GetCode(err)          // e.g., "NOT_FOUND"
```

**Error Phases (Auto-Detected):**
- `PhaseBootstrap` - App initialization
- `PhaseConfig` - Configuration loading
- `PhaseClient` - Client initialization
- `PhaseMiddleware` - Middleware execution
- `PhaseHandler` - HTTP handler
- `PhaseService` - Business logic
- `PhaseRepository` - Data access

### 2.3 DevMode

**Source:** `core/config/devmode.go`

Enable development mode for debugging:

```yaml
# config.yaml
dev_mode: true
```

```bash
# Or via environment
CODO_DEV_MODE=true ./myapp serve

# Or via CLI flag
./myapp serve --dev
```

**DevMode Automatically Enables:**
- `errors.handler.expose_details: true` - Error details in responses
- `errors.handler.expose_stack_traces: true` - Stack traces in responses
- `middleware.auth.dev_mode: true` - Verbose auth logging

**Security Note:** `dev_bypass_auth` must be EXPLICITLY set (never auto-enabled):

```yaml
middleware:
  auth:
    dev_bypass_auth: true  # Must be explicit
    dev_identity:
      id: "dev-user-123"
      traits:
        email: "dev@localhost"
```

### 2.4 Strict Mode

**Source:** `core/config/response.go`

Controls response serialization:

```yaml
# config.yaml
response:
  strict: true  # Include all fields in responses
```

**With strict mode:**
- Empty arrays serialize as `[]` (not `null`)
- Empty strings serialize as `""` (not omitted)
- All defined fields always included

---

## 3. Middleware Creation & Registration

**Source:** `core/middleware/middleware.go`, `core/middleware/orchestrator.go`

### Priority Levels

| Range | Purpose | Examples |
|-------|---------|----------|
| 0-99 | Core (framework-critical) | Recover (0), RequestID (10), ErrorHandler (15) |
| 100-199 | Feature (built-in, configurable) | Logger (100), Auth (105), Timeout (110), CORS (120) |
| 200-299 | Consumer (app-specific) | Your custom middleware |

### Router Scopes

```go
const (
    RouterPublic    Router = 1 << iota  // Port 8081
    RouterProtected                      // Port 8080
    RouterHidden                         // Port 8079
    RouterAll       = RouterPublic | RouterProtected | RouterHidden
)
```

### Creating Custom Middleware

```go
package mymiddleware

import (
    "github.com/labstack/echo/v4"
    "github.com/codoworks/codo-framework/core/middleware"
)

func init() {
    middleware.RegisterMiddleware(&RateLimitMiddleware{
        BaseMiddleware: middleware.NewBaseMiddleware(
            "rate-limit",               // Name
            "middleware.rate_limit",    // ConfigKey
            210,                         // Priority (consumer range)
            middleware.RouterProtected, // Only protected router
        ),
    })
}

type RateLimitMiddleware struct {
    middleware.BaseMiddleware
    requestsPerMinute int
}

// Enabled checks if middleware should be active
func (m *RateLimitMiddleware) Enabled(cfg any) bool {
    rateCfg, ok := cfg.(*RateLimitConfig)
    if !ok {
        return false
    }
    return rateCfg.Enabled
}

// Configure initializes middleware with config
func (m *RateLimitMiddleware) Configure(cfg any) error {
    rateCfg, ok := cfg.(*RateLimitConfig)
    if !ok {
        return nil
    }
    m.requestsPerMinute = rateCfg.RequestsPerMinute
    return nil
}

// Handler returns the actual middleware function
func (m *RateLimitMiddleware) Handler() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Rate limiting logic here
            // ...
            return next(c)
        }
    }
}

type RateLimitConfig struct {
    Enabled           bool `yaml:"enabled"`
    RequestsPerMinute int  `yaml:"requests_per_minute"`
}
```

### Built-in Middleware

| Middleware | Priority | Routers | Purpose |
|------------|----------|---------|---------|
| Recover | 0 | All | Panic recovery with stack traces |
| RequestID | 10 | All | X-Request-ID generation/propagation |
| ErrorHandler | 15 | All | Centralized error handling |
| ContextInit | 20 | All | Context wrapping |
| Logger | 100 | All | Request/response logging |
| Pagination | 102 | All | Pagination parameter extraction (disabled by default) |
| Auth | 105 | Protected | Kratos session validation |
| Timeout | 110 | All | Request timeout enforcement |
| CORS | 120 | All | Cross-origin resource sharing |
| SecurityHeaders | 140 | All | XSS, HSTS, etc. |
| Compression | 150 | All | Gzip responses |

---

## 4. Client Creation & Registration

**Source:** `core/clients/client.go`, `core/clients/registry.go`

### Client Interface

```go
type Client interface {
    Name() string                       // Unique identifier
    ConfigKey() string                  // Config section key
    Configure(cfg any) error            // Receives config section
    Start(ctx context.Context) error    // Initialize/connect
    Stop(ctx context.Context) error     // Graceful shutdown
    Health(ctx context.Context) error   // Health check (nil = healthy)
    Contexts() []Context                // Execution contexts
    DependsOn() []string                // Dependencies
}
```

### Creating a Custom Client

```go
package myclient

import (
    "context"
    "github.com/codoworks/codo-framework/core/clients"
)

const ClientName = "my-client"

func init() {
    clients.MustRegister(New())
}

type Client struct {
    clients.BaseClient
    apiKey string
    url    string
}

func New() *Client {
    return &Client{
        BaseClient: clients.NewBaseClient(
            ClientName,                          // Name
            "my_client",                         // ConfigKey (config.yaml section)
            []clients.Context{clients.ContextAPI}, // When to initialize
            []string{"logger"},                  // Dependencies
        ),
    }
}

type Config struct {
    APIKey string `yaml:"api_key"`
    URL    string `yaml:"url"`
}

func (c *Client) Configure(cfg any) error {
    config, ok := cfg.(*Config)
    if !ok {
        return nil
    }
    c.apiKey = config.APIKey
    c.url = config.URL
    return nil
}

func (c *Client) Start(ctx context.Context) error {
    // Initialize connections, validate config
    return nil
}

func (c *Client) Stop(ctx context.Context) error {
    // Close connections, cleanup
    return nil
}

func (c *Client) Health(ctx context.Context) error {
    // Check if client is healthy
    return nil
}

// Client-specific methods
func (c *Client) DoSomething(ctx context.Context) error {
    // Business logic
    return nil
}
```

### Using Clients

```go
import "github.com/codoworks/codo-framework/core/clients"

// Type-safe retrieval
myClient, err := clients.GetTyped[*myclient.Client]("my-client")
if err != nil {
    return errors.Internal("my-client not available")
}

// Check availability
if clients.Has("my-client") {
    // Client is registered
}

// Get all clients
all := clients.All()

// Shutdown all
clients.ShutdownAll()
```

### Execution Contexts

```go
const (
    ContextAll        Context = "*"          // Always initialized
    ContextAPI        Context = "api"        // HTTP routers
    ContextTasks      Context = "tasks"      // Background tasks
    ContextMigrations Context = "migrations" // Database migrations
    ContextSeeds      Context = "seeds"      // Database seeding
    ContextCLI        Context = "cli"        // CLI commands
)
```

---

## 5. Environment Variables & Configuration

**Source:** `core/config/config.go`, `core/config/env.go`

### Configuration Loading Flow

```
config.yaml (or app.yaml)
    → Parse YAML
    → ApplyDefaults()
    → LoadDotEnv() from .env file
    → applyEnvOverrides() for CODO_* variables
    → Validate()
```

### Config File (config.yaml)

```yaml
service:
  name: my-api
  version: 1.0.0

server:
  public_port: 8081
  protected_port: 8080
  hidden_port: 8079
  read_timeout: 30s
  write_timeout: 30s
  shutdown_grace: 30s
  request_size_limit: 10M

database:
  driver: postgres          # postgres, mysql, sqlite
  host: localhost
  port: 5432
  database: myapp
  username: user
  password: secret
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

auth:
  kratos:
    public_url: http://kratos:4433
    admin_url: http://kratos:4434
  keto:
    read_url: http://keto:4466
    write_url: http://keto:4467

rabbitmq:
  url: amqp://guest:guest@rabbitmq:5672/

logger:
  level: info               # debug, info, warn, error, fatal, panic
  format: json              # json or text

features:
  disabled:
    - rabbitmq              # Disable specific features

middleware:
  logger:
    enabled: true
    skip_paths:
      - /health
  auth:
    enabled: true
    skip_paths:
      - /health
      - /metrics
    cache_enabled: true
    cache_ttl: 15m
  cors:
    enabled: true
    allow_origins:
      - "*"
  pagination:
    enabled: false          # Disabled by default
    default_page_size: 20
    max_page_size: 100

errors:
  handler:
    expose_details: false
    expose_stack_traces: false
  capture:
    stack_trace_on_5xx: true
    stack_trace_on_4xx: false

response:
  strict: false

dev_mode: false
```

### .env File

```bash
# .env (optional, in current working directory)

# Framework variables (CODO_* prefix)
CODO_DB_HOST=localhost
CODO_DB_PASSWORD=supersecret
CODO_DEV_MODE=true
CODO_LOG_LEVEL=debug

# Consumer variables (no prefix needed)
STRIPE_API_KEY=sk_test_xxx
SENDGRID_API_KEY=SG.xxx
KRATOS_ADMIN_URL=http://localhost:4434
```

### Consumer Environment Variables

```go
// In internal/clients/clients.go
func RegisterEnvVars(registry *config.EnvVarRegistry) error {
    return registry.RegisterMany([]config.EnvVarDescriptor{
        {
            Name:        "STRIPE_API_KEY",
            Type:        config.EnvTypeString,
            Required:    true,
            Sensitive:   true,
            Group:       "stripe",
            Description: "Stripe secret API key",
        },
        {
            Name:        "WEBHOOK_URL",
            Type:        config.EnvTypeURL,
            Required:    false,
            Default:     "http://localhost:3000/webhook",
            Group:       "webhooks",
            Description: "Webhook callback URL",
        },
        {
            Name:        "CACHE_TTL",
            Type:        config.EnvTypeDuration,
            Required:    false,
            Default:     "5m",
            Group:       "cache",
            Description: "Cache TTL duration",
        },
    })
}

// Accessing values
func Initialize(cfg *config.Config) error {
    stripeVars := cfg.EnvRegistry.GetGroup("stripe")
    apiKey := stripeVars["STRIPE_API_KEY"].String()

    cacheVars := cfg.EnvRegistry.GetGroup("cache")
    ttl := cacheVars["CACHE_TTL"].Duration()

    // Type-safe access
    value := cfg.EnvRegistry.Get("WEBHOOK_URL")
    if value.IsSet {
        url := value.URL()
    }
}
```

### Environment Variable Types

| Type | Description | Example |
|------|-------------|---------|
| `EnvTypeString` | Regular string | `"hello"` |
| `EnvTypeInt` | Integer | `42` |
| `EnvTypeBool` | Boolean | `true`, `false`, `1`, `0` |
| `EnvTypeDuration` | Duration | `"30s"`, `"5m"`, `"1h"` |
| `EnvTypeFloat` | Float64 | `3.14` |
| `EnvTypeURL` | URL with validation | `"https://example.com"` |

---

## 6. Tasks

**Source:** `.claude/specs/11-tasks.md`

Tasks are one-time executable jobs run via CLI.

### Task Interface

```go
type Task interface {
    Group() string                    // e.g., "user"
    Name() string                     // e.g., "promote-admin"
    Description() string              // Human-readable description
    Parameters() []Parameter          // CLI parameters
    Interactive() bool                // Uses prompts/confirmations
    Run(ctx context.Context, r *Runner) error
}
```

### Creating a Task

```go
// pkg/tasks/user/promote_admin.go
package user

import (
    "context"
    "fmt"
    "github.com/codoworks/codo-framework/core/clients"
    "github.com/codoworks/codo-framework/core/db"
    "github.com/codoworks/codo-framework/core/tasks"
)

func init() {
    tasks.Register(&PromoteAdminTask{})
}

type PromoteAdminTask struct{}

func (t *PromoteAdminTask) Group() string       { return "user" }
func (t *PromoteAdminTask) Name() string        { return "promote-admin" }
func (t *PromoteAdminTask) Description() string { return "Promote a user to admin role" }
func (t *PromoteAdminTask) Interactive() bool   { return true }

func (t *PromoteAdminTask) Parameters() []tasks.Parameter {
    return []tasks.Parameter{
        {
            Name:        "user-id",
            Short:       "u",
            Type:        tasks.ParamTypeString,
            Required:    true,
            Description: "The user's UUID to promote",
            ValidateTag: "uuid4",
        },
        {
            Name:        "reason",
            Short:       "r",
            Type:        tasks.ParamTypeString,
            Required:    true,
            Description: "Reason for the promotion",
            ValidateTag: "min=10,max=500",
        },
    }
}

func (t *PromoteAdminTask) Run(ctx context.Context, r *tasks.Runner) error {
    userID := r.Params().String("user-id")
    reason := r.Params().String("reason")

    r.Info(fmt.Sprintf("Promoting user: %s", userID))
    r.Info(fmt.Sprintf("Reason: %s", reason))

    if !r.Confirm("Proceed with promotion?") {
        r.Info("Operation cancelled")
        return nil
    }

    // Get database client
    dbClient, err := clients.GetTyped[*db.Client]("db")
    if err != nil {
        return fmt.Errorf("database not available: %w", err)
    }

    // Perform the promotion
    _, err = dbClient.DB().ExecContext(ctx,
        "UPDATE users SET role = 'admin' WHERE id = $1",
        userID,
    )
    if err != nil {
        r.Error(err)
        return err
    }

    r.Success("User promoted successfully")
    return nil
}
```

### Runner Methods

```go
// Output
r.Info(msg)          // Informational message
r.Success(msg)       // Success message
r.Warning(msg)       // Warning message
r.Error(err)         // Error message
r.Debug(msg)         // Debug (dev mode only)

// Interactive
r.Confirm(msg) bool                              // Yes/no confirmation
r.ConfirmDanger(msg, confirmPhrase) bool         // Type phrase to confirm
r.Prompt(msg) string                             // Text input
r.PromptWithDefault(msg, default) string         // With default value
r.PromptRequired(msg) (string, error)            // Re-prompt if empty
r.PromptInt(msg) (int, error)                    // Integer input
r.Select(msg, options) int                       // Selection from list

// Parameters
r.Params().String(name) string
r.Params().Int(name) int
r.Params().Bool(name) bool
r.Params().StringSlice(name) []string

// Modes
r.IsDevMode() bool
r.IsYesMode() bool  // --yes flag skips confirmations
```

### CLI Usage

```bash
# List all tasks
./myapp task list

# Run a task with parameters
./myapp task run user:promote-admin --user-id=abc-123 --reason="Performance review"

# Run with interactive prompts (missing params)
./myapp task run user:promote-admin

# Skip confirmations
./myapp task run user:promote-admin --user-id=abc-123 --reason="Urgent" --yes
```

---

## 7. Models, Migrations, Seeds

**Source:** `core/db/model.go`, `core/db/repository.go`, `.claude/specs/09-migrations.md`

### Model Definition

```go
// pkg/models/user.go
package models

import "github.com/codoworks/codo-framework/core/db"

type User struct {
    db.Model                          // ID, CreatedAt, UpdatedAt, DeletedAt
    Email        string  `db:"email"`
    Name         string  `db:"name"`
    PasswordHash string  `db:"password_hash"`
    Role         string  `db:"role"`
    IsActive     bool    `db:"is_active"`
    GroupID      *string `db:"group_id"` // Nullable foreign key
}

func (u *User) TableName() string {
    return "users"
}

// Optional hooks
func (u *User) BeforeCreate(ctx context.Context) error {
    // Generate ID, set defaults
    return nil
}

func (u *User) AfterCreate(ctx context.Context) error {
    // Post-creation logic
    return nil
}

func (u *User) BeforeUpdate(ctx context.Context) error {
    // Pre-update validation
    return nil
}

func (u *User) BeforeDelete(ctx context.Context) error {
    // Cascade logic
    return nil
}
```

### Model Base Fields

```go
type Model struct {
    ID        string     `db:"id"`         // UUID, auto-generated
    CreatedAt time.Time  `db:"created_at"` // Auto-set on create
    UpdatedAt time.Time  `db:"updated_at"` // Auto-updated
    DeletedAt *time.Time `db:"deleted_at"` // Soft delete
}

// Methods
model.IsNew()        // Not yet persisted
model.IsPersisted()  // Has ID and CreatedAt
model.IsDeleted()    // Soft deleted
model.MarkDeleted()  // Set DeletedAt
model.Restore()      // Clear DeletedAt
```

### Migration Definition

```go
// pkg/migrations/20251220120000_create_users_table.go
package migrations

import (
    "context"
    "github.com/jmoiron/sqlx"
    "github.com/codoworks/codo-framework/core/db/migrations"
)

func init() {
    migrations.Register(&CreateUsersTable{})
}

type CreateUsersTable struct{}

func (m CreateUsersTable) Version() string { return "20251220120000" }
func (m CreateUsersTable) Name() string    { return "create_users_table" }

func (m CreateUsersTable) Migrate(ctx context.Context, tx *sqlx.Tx) error {
    _, err := tx.ExecContext(ctx, `
        CREATE TABLE users (
            id VARCHAR(36) PRIMARY KEY,
            email VARCHAR(255) UNIQUE NOT NULL,
            name VARCHAR(255) NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            role VARCHAR(50) NOT NULL DEFAULT 'user',
            is_active BOOLEAN NOT NULL DEFAULT true,
            group_id VARCHAR(36),
            created_at TIMESTAMP NOT NULL,
            updated_at TIMESTAMP NOT NULL,
            deleted_at TIMESTAMP NULL,
            FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL
        );
        CREATE INDEX idx_users_email ON users(email);
        CREATE INDEX idx_users_deleted_at ON users(deleted_at);
    `)
    return err
}

func (m CreateUsersTable) Rollback(ctx context.Context, tx *sqlx.Tx) error {
    _, err := tx.ExecContext(ctx, `
        DROP INDEX IF EXISTS idx_users_deleted_at;
        DROP INDEX IF EXISTS idx_users_email;
        DROP TABLE IF EXISTS users;
    `)
    return err
}
```

### Seed Definition

```go
// pkg/seeds/20251220130000_admin_user.go
package seeds

import (
    "context"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/codoworks/codo-framework/core/db/seeds"
)

func init() {
    seeds.Register(&AdminUserSeed{})
}

type AdminUserSeed struct{}

func (s AdminUserSeed) Version() string { return "20251220130000" }
func (s AdminUserSeed) Name() string    { return "admin_user" }

func (s AdminUserSeed) Environments() []seeds.Environment {
    return []seeds.Environment{seeds.EnvDevelopment, seeds.EnvTest}
}

func (s AdminUserSeed) Seed(ctx context.Context, tx *sqlx.Tx) error {
    _, err := tx.ExecContext(ctx, `
        INSERT INTO users (id, email, name, password_hash, role, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
        ON CONFLICT (email) DO NOTHING
    `, uuid.NewString(), "admin@example.com", "Admin User", "$2a$10$...", "admin")
    return err
}

func (s AdminUserSeed) Rollback(ctx context.Context, tx *sqlx.Tx) error {
    _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE email = $1`, "admin@example.com")
    return err
}
```

### Migration/Seed Registration

```go
// pkg/migrations/migrations.go
package migrations

import "github.com/codoworks/codo-framework/core/db/migrations"

func All() []*migrations.Migration {
    return []*migrations.Migration{
        // Return in order
    }
}

func AddToRunner(runner *migrations.Runner) {
    runner.Add(All()...)
}

// pkg/seeds/seeds.go
package seeds

import "github.com/codoworks/codo-framework/core/db/seeds"

func All() []*seeds.Seed {
    return []*seeds.Seed{
        // Return seeds
    }
}

func AddToSeeder(seeder *seeds.Seeder) {
    seeder.Add(All()...)
}
```

---

## 8. Forms, Services, Handlers

### Forms (DTOs)

```go
// pkg/forms/user_forms.go
package forms

import (
    "time"
    "myapp/pkg/models"
)

// Request form for creating users
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email,max=255"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8,max=72"`
    Role     string `json:"role" validate:"omitempty,oneof=user admin"`
}

func (f *CreateUserRequest) ToModel() *models.User {
    return &models.User{
        Email: f.Email,
        Name:  f.Name,
        Role:  f.defaultRole(),
    }
}

func (f *CreateUserRequest) defaultRole() string {
    if f.Role == "" {
        return "user"
    }
    return f.Role
}

// Update form with optional fields (pointer = optional)
type UpdateUserRequest struct {
    Name     *string `json:"name" validate:"omitempty,min=2,max=100"`
    Email    *string `json:"email" validate:"omitempty,email,max=255"`
    Role     *string `json:"role" validate:"omitempty,oneof=user admin"`
    IsActive *bool   `json:"is_active"`
}

func (f *UpdateUserRequest) ApplyTo(user *models.User) {
    if f.Name != nil {
        user.Name = *f.Name
    }
    if f.Email != nil {
        user.Email = *f.Email
    }
    if f.Role != nil {
        user.Role = *f.Role
    }
    if f.IsActive != nil {
        user.IsActive = *f.IsActive
    }
}

// Response form (what API returns)
type UserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    Role      string    `json:"role"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    // Note: PasswordHash NEVER included
}

func NewUserResponse(user *models.User) *UserResponse {
    return &UserResponse{
        ID:        user.ID,
        Email:     user.Email,
        Name:      user.Name,
        Role:      user.Role,
        IsActive:  user.IsActive,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }
}

func NewUserListResponse(users []*models.User) []*UserResponse {
    responses := make([]*UserResponse, len(users))
    for i, user := range users {
        responses[i] = NewUserResponse(user)
    }
    return responses
}
```

### Services (Business Logic)

```go
// pkg/services/user_service.go
package services

import (
    "context"
    "golang.org/x/crypto/bcrypt"

    "github.com/codoworks/codo-framework/core/db"
    "github.com/codoworks/codo-framework/core/errors"

    "myapp/pkg/forms"
    "myapp/pkg/models"
)

type UserService struct {
    repo *db.Repository[*models.User]
}

func NewUserService(dbClient *db.Client) *UserService {
    return &UserService{
        repo: db.NewRepository[*models.User](dbClient),
    }
}

func (s *UserService) Create(ctx context.Context, form *forms.CreateUserRequest) (*models.User, error) {
    // Check for duplicate email
    existing, _ := s.repo.FindOne(ctx, db.Where("email = ?", form.Email))
    if existing != nil {
        return nil, errors.Conflict("Email already in use").
            WithField("email", "VALIDATION.UNIQUE", "Email is already registered")
    }

    // Hash password (business logic stays in service)
    hash, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.Internal("Failed to hash password").WithCause(err)
    }

    // Create model
    user := form.ToModel()
    user.PasswordHash = string(hash)
    user.IsActive = true

    // Save
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (s *UserService) FindByID(ctx context.Context, id string) (*models.User, error) {
    record, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    return record.Model(), nil
}

func (s *UserService) Update(ctx context.Context, id string, form *forms.UpdateUserRequest) (*models.User, error) {
    record, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    user := record.Model()

    // Check email uniqueness if changing
    if form.Email != nil && *form.Email != user.Email {
        existing, _ := s.repo.FindOne(ctx,
            db.Where("email = ? AND id != ?", *form.Email, id))
        if existing != nil {
            return nil, errors.Conflict("Email already in use").
                WithField("email", "VALIDATION.UNIQUE", "Email is already registered")
        }
    }

    form.ApplyTo(user)

    if err := s.repo.Update(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
    record, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return err
    }
    return s.repo.Delete(ctx, record.Model())
}

func (s *UserService) FindAll(ctx context.Context, opts ...db.QueryOption) ([]*models.User, error) {
    records, err := s.repo.FindAll(ctx, opts...)
    if err != nil {
        return nil, err
    }

    users := make([]*models.User, len(records))
    for i, r := range records {
        users[i] = r.Model()
    }
    return users, nil
}

func (s *UserService) Count(ctx context.Context, opts ...db.QueryOption) (int64, error) {
    return s.repo.Count(ctx, opts...)
}
```

### Handlers (HTTP Layer)

```go
// pkg/handlers/user_handler.go
package handlers

import (
    "github.com/labstack/echo/v4"

    "github.com/codoworks/codo-framework/core/db"
    "github.com/codoworks/codo-framework/core/errors"
    "github.com/codoworks/codo-framework/core/forms"
    "github.com/codoworks/codo-framework/core/http"
    "github.com/codoworks/codo-framework/core/pagination"

    userforms "myapp/pkg/forms"
    "myapp/pkg/services"
)

type UserHandler struct {
    service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
    return &UserHandler{service: service}
}

func (h *UserHandler) Prefix() string {
    return "/api/v1/users"
}

func (h *UserHandler) Scope() http.RouterScope {
    return http.ScopeProtected
}

func (h *UserHandler) Middlewares() []echo.MiddlewareFunc {
    return nil
}

func (h *UserHandler) Initialize() error {
    return nil
}

func (h *UserHandler) Routes(g *echo.Group) {
    g.GET("", http.WrapHandler(h.List))
    g.POST("", http.WrapHandler(h.Create))
    g.GET("/:id", http.WrapHandler(h.Get))
    g.PUT("/:id", http.WrapHandler(h.Update))
    g.DELETE("/:id", http.WrapHandler(h.Delete))
}

func (h *UserHandler) List(c *http.Context) error {
    ctx := c.Request().Context()

    // Get pagination (from middleware or defaults)
    params := pagination.GetOrDefault(c.Context, 1, 20)

    opts := []db.QueryOption{
        db.Limit(params.PerPage),
        db.Offset(params.Offset),
        db.OrderByDesc("created_at"),
    }

    // Optional filtering
    if role := c.QueryParam("role"); role != "" {
        opts = append(opts, db.Where("role = ?", role))
    }

    total, err := h.service.Count(ctx)
    if err != nil {
        return err
    }

    users, err := h.service.FindAll(ctx, opts...)
    if err != nil {
        return err
    }

    items := userforms.NewUserListResponse(users)
    response := forms.NewListResponse(items, total, params.Page, params.PerPage)

    return c.Success(response)
}

func (h *UserHandler) Create(c *http.Context) error {
    ctx := c.Request().Context()

    var form userforms.CreateUserRequest
    if err := c.BindAndValidate(&form); err != nil {
        return err
    }

    user, err := h.service.Create(ctx, &form)
    if err != nil {
        return err
    }

    return c.Created(userforms.NewUserResponse(user))
}

func (h *UserHandler) Get(c *http.Context) error {
    ctx := c.Request().Context()

    id, err := c.ParamUUID("id")
    if err != nil {
        return err
    }

    user, err := h.service.FindByID(ctx, id)
    if err != nil {
        return err
    }

    return c.Success(userforms.NewUserResponse(user))
}

func (h *UserHandler) Update(c *http.Context) error {
    ctx := c.Request().Context()

    id, err := c.ParamUUID("id")
    if err != nil {
        return err
    }

    var form userforms.UpdateUserRequest
    if err := c.BindAndValidate(&form); err != nil {
        return err
    }

    user, err := h.service.Update(ctx, id, &form)
    if err != nil {
        return err
    }

    return c.Success(userforms.NewUserResponse(user))
}

func (h *UserHandler) Delete(c *http.Context) error {
    ctx := c.Request().Context()

    id, err := c.ParamUUID("id")
    if err != nil {
        return err
    }

    if err := h.service.Delete(ctx, id); err != nil {
        return err
    }

    return c.NoContent()
}
```

---

## 9. Logging

**Source:** `clients/logger/logger.go`

### Getting the Logger

```go
import (
    "github.com/codoworks/codo-framework/clients/logger"
    "github.com/codoworks/codo-framework/core/clients"
)

// Get logger client
log, err := clients.GetTyped[*logger.Logger](logger.ClientName)
if err != nil {
    // Handle error
}
```

### Basic Logging

```go
log.Debug("Debug message")
log.Info("Info message")
log.Warn("Warning message")
log.Error("Error message")
log.Fatal("Fatal message")  // Exits program
log.Panic("Panic message")  // Panics

// Formatted
log.Debugf("User %s logged in", userID)
log.Infof("Processing %d items", count)
log.Errorf("Failed to process: %v", err)
```

### Structured Logging

```go
log.WithField("user_id", userID).Info("User logged in")

log.WithFields(map[string]any{
    "user_id":    userID,
    "session_id": sessionID,
    "ip":         remoteAddr,
}).Info("Authentication successful")

log.WithError(err).Error("Failed to process request")

log.WithFields(map[string]any{
    "operation": "create_user",
    "email":     email,
}).WithError(err).Error("User creation failed")
```

### Log Levels

| Level | Purpose |
|-------|---------|
| `debug` | Detailed debugging information |
| `info` | General operational information |
| `warn` | Warning conditions |
| `error` | Error conditions |
| `fatal` | Fatal conditions (exits program) |
| `panic` | Panic conditions (panics) |

### Configuration

```yaml
logger:
  level: info    # debug, info, warn, error, fatal, panic
  format: json   # json or text
```

---

## 10. Advanced Features

### 10.1 Multi-Router Architecture

Three independent HTTP routers for Kubernetes ingress separation:

| Router | Port | Purpose | Auth |
|--------|------|---------|------|
| Public | 8081 | Health checks, public endpoints | None |
| Protected | 8080 | Main API, user-facing | Kratos session |
| Hidden | 8079 | Admin, service-to-service | Custom |

```go
// Handler targets specific router via Scope()
func (h *HealthHandler) Scope() http.RouterScope {
    return http.ScopePublic  // Port 8081
}

func (h *UserHandler) Scope() http.RouterScope {
    return http.ScopeProtected  // Port 8080
}

func (h *AdminHandler) Scope() http.RouterScope {
    return http.ScopeHidden  // Port 8079
}

// Or target multiple
func (h *MetricsHandler) Scope() http.RouterScope {
    return http.ScopePublic | http.ScopeHidden
}
```

### 10.2 Bootstrap Modes

```go
const (
    HTTPServer      // Full multi-router (serve command)
    HTTPRouter      // Single router (start public/protected/hidden)
    WorkerDaemon    // Background workers (no HTTP)
    RouteInspector  // Route introspection (info routes)
    ConfigInspector // Config inspection (info env)
)
```

**Single Router (Kubernetes deployment):**
```bash
# Deploy each router as separate pod
./myapp start public     # Public pod
./myapp start protected  # Protected pod
./myapp start hidden     # Hidden pod
```

### 10.3 Repository Pattern (sqlx)

```go
repo := db.NewRepository[*models.User](dbClient)

// Query options
repo.FindAll(ctx,
    db.Where("status = ?", "active"),
    db.WhereEq("role", "admin"),
    db.WhereIn("status", "active", "pending"),
    db.WhereLike("name", "%john%"),
    db.WhereBetween("age", 18, 65),
    db.WhereNull("deleted_at"),
    db.WhereNotNull("email"),
    db.OrderByAsc("created_at"),
    db.OrderByDesc("updated_at"),
    db.Limit(10),
    db.Offset(20),
    db.WithDeleted(),  // Include soft-deleted
)

// Record wrapper for change tracking
record := repo.New()
record.Model().Name = "John"
record.Save(ctx)  // Creates

record.Model().Name = "Jane"
record.Save(ctx)  // Updates

record.Delete(ctx)  // Soft delete
```

### 10.4 Feature Toggles

```yaml
features:
  disabled:
    - rabbitmq
    - redis
    - cors
```

```bash
# Via environment
export DISABLE_FEATURES="rabbitmq,redis"
```

```go
if cfg.Features.IsEnabled("rabbitmq") {
    // Initialize RabbitMQ
}

if cfg.Features.IsDisabled("cors") {
    // Skip CORS setup
}
```

### 10.5 Health Checks

Built-in endpoints:
- `GET /health/alive` - Always returns 200 (liveness probe)
- `GET /health/ready` - Checks all components (readiness probe)

```go
// Register custom health checker
http.RegisterNamedHealthChecker("database", func() error {
    return dbClient.Ping(ctx)
})

http.RegisterNamedHealthChecker("cache", func() error {
    return redisClient.Ping(ctx)
})
```

### 10.6 Cursor Pagination

```yaml
middleware:
  pagination:
    enabled: true
    default_type: cursor  # or offset
```

```go
func (h *UserHandler) List(c *http.Context) error {
    params := pagination.Get(c)

    if params.IsCursor() {
        // Cursor-based
        users, err := h.service.FindAll(ctx,
            db.Where("id > ?", params.Cursor),
            db.Limit(params.PerPage),
        )
        // Set cursor for next page
        if len(users) > 0 {
            pagination.SetCursorMeta(c.Context, users[len(users)-1].ID, len(users) == params.PerPage)
        }
    } else {
        // Offset-based
        users, err := h.service.FindAll(ctx,
            db.Limit(params.PerPage),
            db.Offset(params.Offset),
        )
        pagination.SetMeta(c.Context, len(users), total)
    }
}
```

### 10.7 Partial Success Responses

```go
func (h *UserHandler) BatchDelete(c *http.Context) error {
    var form BatchDeleteRequest
    if err := c.BindAndValidate(&form); err != nil {
        return err
    }

    deleted := 0
    for _, id := range form.IDs {
        err := h.service.Delete(ctx, id)
        if err != nil {
            // Add warning, continue processing
            c.AddWarning("DELETE_FAILED", fmt.Sprintf("Failed to delete %s: %v", id, err))
            continue
        }
        deleted++
    }

    if deleted == 0 && c.HasWarnings() {
        return errors.BadRequest("No items were deleted")
    }

    return c.Success(map[string]int{
        "deleted": deleted,
        "total":   len(form.IDs),
    })
}
```

Response:
```json
{
  "code": "SUCCESS",
  "payload": {"deleted": 3, "total": 5},
  "warnings": [
    {"code": "DELETE_FAILED", "message": "Failed to delete abc-123: not found"},
    {"code": "DELETE_FAILED", "message": "Failed to delete def-456: permission denied"}
  ]
}
```

### 10.8 Authentication Context

```go
import "github.com/codoworks/codo-framework/core/auth"

func (h *UserHandler) GetMe(c *http.Context) error {
    // Get authenticated identity
    identity, err := auth.GetIdentity(c)
    if err != nil {
        return errors.Unauthorized("Not authenticated")
    }

    userID := identity.ID
    email := identity.Email()
    name := identity.Name()

    // Access custom traits
    customValue, ok := identity.GetTrait("custom_key")
    strValue := identity.GetTraitString("key")
    boolValue := identity.GetTraitBool("key")

    // Or panic if required
    identity := auth.MustGetIdentity(c)
}
```

---

## Reference: Key File Locations

| Feature | Path |
|---------|------|
| Bootstrap | `core/app/bootstrap.go` |
| Config | `core/config/config.go` |
| Validation | `core/forms/validation.go` |
| Errors | `core/errors/errors.go` |
| Middleware | `core/middleware/orchestrator.go` |
| Clients | `core/clients/registry.go` |
| Repository | `core/db/repository.go` |
| Model | `core/db/model.go` |
| Auth | `core/auth/identity.go` |
| Pagination | `core/middleware/pagination/pagination.go` |
| Logger | `clients/logger/logger.go` |
| Specs | `.claude/specs/` |

---

## See Also

- [CLAUDE.md](../CLAUDE.md) - Framework development guidelines
- [Specs](./.claude/specs/) - Detailed specifications for each component
- [Examples](../examples/) - Working code examples
