package recover

import (
	"fmt"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/errors"
	httpPkg "github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/core/middleware"
)

func init() {
	middleware.RegisterMiddleware(&RecoverMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"recover",
			"middleware.recover",
			middleware.PriorityRecover,
			middleware.RouterAll,
		),
	})
}

// RecoverMiddleware handles panic recovery
type RecoverMiddleware struct {
	middleware.BaseMiddleware
	logger *logrus.Logger
	cfg    *config.Config
}

// Enabled always returns true for core middleware
func (m *RecoverMiddleware) Enabled(cfg any) bool {
	return true // Core middleware is always enabled
}

// Configure initializes the middleware with logger from client registry
func (m *RecoverMiddleware) Configure(cfg any) error {
	// Get logger from client registry
	loggerClient, err := clients.GetTyped[*logger.Logger]("logger")
	if err != nil {
		return fmt.Errorf("failed to get logger client: %w", err)
	}

	m.logger = loggerClient.GetLogger()

	// Store config if provided (for dev mode check)
	if configVal, ok := cfg.(*config.Config); ok {
		m.cfg = configVal
	}

	return nil
}

// Handler returns the panic recovery middleware function
func (m *RecoverMiddleware) Handler() echo.MiddlewareFunc {
	log := m.logger
	if log == nil {
		log = logrus.StandardLogger()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Convert panic value to error
					var err error
					switch x := r.(type) {
					case string:
						err = fmt.Errorf("panic: %s", x)
					case error:
						err = x
					default:
						err = fmt.Errorf("panic: %v", r)
					}

					stack := debug.Stack()

					// Create framework error from panic
					fwkErr := errors.WrapInternal(err, "Panic recovered in HTTP handler").
						WithPhase(errors.PhaseHandler).
						WithDetail("panic", true).
						WithDetail("stack", string(stack))

					// Enrich with request context
					fwkErr.RequestCtx = &errors.RequestContext{
						RequestID:  c.Response().Header().Get(echo.HeaderXRequestID),
						Method:     c.Request().Method,
						Path:       c.Request().URL.Path,
						RemoteAddr: c.RealIP(),
					}

					// Log panic with full details
					log.WithFields(logrus.Fields{
						"error":     err.Error(),
						"stack":     string(stack),
						"requestId": fwkErr.RequestCtx.RequestID,
						"path":      fwkErr.RequestCtx.Path,
					}).Error("Panic recovered")

					// Render response using framework error
					// Note: Error handler middleware won't run after panic, so we handle it here
					resp := httpPkg.ErrorResponse(fwkErr)
					c.JSON(resp.HTTPStatus, resp)
				}
			}()

			return next(c)
		}
	}
}
