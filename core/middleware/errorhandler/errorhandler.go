package errorhandler

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/errors"
	"github.com/codoworks/codo-framework/core/middleware"
	httpPkg "github.com/codoworks/codo-framework/core/http"
)

func init() {
	middleware.RegisterMiddleware(&ErrorHandlerMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"errorhandler",
			"middleware.errorhandler",
			middleware.PriorityErrorHandler, // 15 - after recover (0) and requestid (10)
			middleware.RouterAll,
		),
	})
}

// ErrorHandlerMiddleware provides centralized error handling for HTTP requests
type ErrorHandlerMiddleware struct {
	middleware.BaseMiddleware
	logger  *logrus.Logger
	devMode bool
}

// Enabled always returns true for core middleware
func (m *ErrorHandlerMiddleware) Enabled(cfg any) bool {
	return true // Core middleware is always enabled
}

// Configure initializes the middleware with logger and config
func (m *ErrorHandlerMiddleware) Configure(cfg any) error {
	// Get logger from client registry
	loggerClient, err := clients.GetTyped[*logger.Logger]("logger")
	if err != nil {
		return err
	}
	m.logger = loggerClient.GetLogger()

	// Store dev mode flag
	if configVal, ok := cfg.(*config.Config); ok {
		m.devMode = configVal.IsDevMode()
	}

	return nil
}

// Handler returns the error handling middleware function
func (m *ErrorHandlerMiddleware) Handler() echo.MiddlewareFunc {
	log := m.logger
	if log == nil {
		log = logrus.StandardLogger()
	}

	devMode := m.devMode

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Execute handler
			err := next(c)
			if err == nil {
				return nil
			}

			// Map to framework error
			fwkErr := errors.MapError(err)

			// Enrich with request context
			fwkErr.RequestCtx = &errors.RequestContext{
				RequestID:  c.Response().Header().Get(echo.HeaderXRequestID),
				Method:     c.Request().Method,
				Path:       c.Request().URL.Path,
				RemoteAddr: c.RealIP(),
			}

			// Log the error with appropriate level
			logError(log, fwkErr, devMode)

			// Set Retry-After header if error is retryable with a delay
			if fwkErr.Retryable && fwkErr.RetryAfter > 0 {
				c.Response().Header().Set("Retry-After", strconv.Itoa(int(fwkErr.RetryAfter.Seconds())))
			}

			// Render HTTP response
			resp := httpPkg.ErrorResponse(fwkErr)
			return c.JSON(resp.HTTPStatus, resp)
		}
	}
}

// logError logs the error with appropriate fields and level
func logError(log *logrus.Logger, err *errors.Error, devMode bool) {
	fields := logrus.Fields{
		"code":   err.Code,
		"status": err.HTTPStatus,
	}

	// Add request context if present (nil-safe)
	if err.RequestCtx != nil {
		fields["requestId"] = err.RequestCtx.RequestID
		fields["path"] = err.RequestCtx.Path
		fields["method"] = err.RequestCtx.Method
		if err.RequestCtx.UserID != "" {
			fields["userId"] = err.RequestCtx.UserID
		}
	}

	// Add phase if present
	if err.Phase != "" {
		fields["phase"] = err.Phase
	}

	// Add caller location if present
	if err.Caller != nil {
		fields["location"] = err.Caller.File + ":" + strconv.Itoa(err.Caller.Line)
		fields["function"] = err.Caller.Function
	}

	// In dev mode, add more details
	if devMode {
		if len(err.Details) > 0 {
			fields["details"] = err.Details
		}
		if len(err.StackTrace) > 0 {
			fields["stackTrace"] = err.StackTrace
		}
	}

	// Log at appropriate level based on error's LogLevel (if set) or HTTP status
	entry := log.WithFields(fields)
	msg := err.Message
	if err.Cause != nil {
		msg = err.Error() // Include cause in log message
	}

	// Use LogLevel from mapper if set, otherwise fall back to HTTP status
	switch err.LogLevel {
	case errors.LogLevelDebug:
		entry.Debugf("HTTP %d: %s", err.HTTPStatus, msg)
	case errors.LogLevelInfo:
		entry.Infof("HTTP %d: %s", err.HTTPStatus, msg)
	case errors.LogLevelWarn:
		entry.Warnf("HTTP %d: %s", err.HTTPStatus, msg)
	case errors.LogLevelError:
		entry.Errorf("HTTP %d: %s", err.HTTPStatus, msg)
	default:
		// Fall back to HTTP status-based log level
		switch {
		case err.HTTPStatus >= 500:
			entry.Errorf("HTTP %d: %s", err.HTTPStatus, msg)
		case err.HTTPStatus >= 400:
			entry.Warnf("HTTP %d: %s", err.HTTPStatus, msg)
		default:
			entry.Infof("HTTP %d: %s", err.HTTPStatus, msg)
		}
	}
}
