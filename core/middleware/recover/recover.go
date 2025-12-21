package recover

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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

	devMode := false
	if m.cfg != nil {
		devMode = m.cfg.IsDevMode()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					stack := debug.Stack()

					log.WithFields(logrus.Fields{
						"error": err.Error(),
						"stack": string(stack),
						"path":  c.Request().URL.Path,
					}).Error("Panic recovered")

					response := map[string]interface{}{
						"code":    "INTERNAL_ERROR",
						"message": "An unexpected error occurred",
					}

					if devMode {
						response["error"] = err.Error()
						response["stack"] = string(stack)
					}

					c.JSON(http.StatusInternalServerError, response)
				}
			}()

			return next(c)
		}
	}
}
