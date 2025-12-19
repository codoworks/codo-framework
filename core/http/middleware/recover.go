package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// RecoverConfig holds recover middleware configuration
type RecoverConfig struct {
	Logger  *logrus.Logger
	DevMode bool
}

// DefaultRecoverConfig returns a default recover configuration
func DefaultRecoverConfig() *RecoverConfig {
	return &RecoverConfig{
		Logger:  logrus.StandardLogger(),
		DevMode: false,
	}
}

// Recover returns a panic recovery middleware
func Recover(cfg *RecoverConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = DefaultRecoverConfig()
	}

	log := cfg.Logger
	if log == nil {
		log = logrus.StandardLogger()
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

					if cfg.DevMode {
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

// DefaultRecover returns a recover middleware with default settings
func DefaultRecover() echo.MiddlewareFunc {
	return Recover(nil)
}

// RecoverWithLogger returns a recover middleware with a custom logger
func RecoverWithLogger(log *logrus.Logger) echo.MiddlewareFunc {
	return Recover(&RecoverConfig{Logger: log})
}
