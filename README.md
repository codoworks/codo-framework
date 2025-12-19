# Codo Framework

A production-ready Go backend framework with multi-router HTTP architecture, database abstraction, and authentication integrations.

## Installation

```bash
go get github.com/codoworks/codo-framework
```

## Quick Start

```go
package main

import (
    "github.com/codoworks/codo-framework/core/config"
    "github.com/codoworks/codo-framework/core/db"
    "github.com/codoworks/codo-framework/core/http"
)

func main() {
    // Load configuration
    cfg, _ := config.Load()

    // Connect to database
    dbClient, _ := db.NewClient(cfg.Database)

    // Register handlers
    http.RegisterHandler(NewUserHandler(dbClient))

    // Start server
    router := http.NewRouter(http.ScopeProtected, cfg)
    router.Start()
}
```

## Multi-Router Architecture

Three independent routers for Kubernetes ingress separation:

| Router | Port | Scope | Purpose |
|--------|------|-------|---------|
| Public | 8081 | `http.ScopePublic` | Health checks, public endpoints |
| Protected | 8080 | `http.ScopeProtected` | User-facing API (requires auth) |
| Hidden | 8079 | `http.ScopeHidden` | Admin/internal endpoints |

## Core Packages

| Package | Description |
|---------|-------------|
| `core/config` | Configuration loading (env, files) |
| `core/db` | Database client, repository pattern, migrations |
| `core/http` | Router, handlers, middleware, context |
| `core/auth` | Ory Kratos/Keto integration |
| `testutil` | Testing utilities for handlers |

## Creating a Handler

```go
type UserHandler struct {
    service *UserService
}

func (h *UserHandler) Prefix() string           { return "/api/v1/users" }
func (h *UserHandler) Scope() http.RouterScope  { return http.ScopeProtected }
func (h *UserHandler) Middlewares() []echo.MiddlewareFunc { return nil }
func (h *UserHandler) Initialize() error        { return nil }

func (h *UserHandler) Routes(g *echo.Group) {
    g.GET("", http.WrapHandler(h.List))
    g.POST("", http.WrapHandler(h.Create))
    g.GET("/:id", http.WrapHandler(h.Get))
}
```

## CLI Commands

```bash
codo serve              # Start all servers
codo start public       # Start public server only
codo start protected    # Start protected server only
codo db migrate         # Run migrations
codo db rollback        # Rollback migrations
codo db seed            # Run seeds
codo info routes        # Show registered routes
codo info env           # Show environment info
```

## Documentation

Full documentation: *Coming soon*

## License

MIT
