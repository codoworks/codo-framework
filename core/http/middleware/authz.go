package middleware

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/codoworks/codo-framework/core/auth"
)

// PermissionChecker checks permissions
type PermissionChecker interface {
	CheckPermission(ctx context.Context, subject, relation, namespace, object string) (bool, error)
}

// AuthzConfig holds authorization middleware configuration
type AuthzConfig struct {
	Keto       PermissionChecker
	Namespace  string
	Relation   string
	ObjectFunc func(c echo.Context) string
}

// Authz returns an authorization middleware
func Authz(cfg *AuthzConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get identity from context
			identity, err := auth.GetIdentity(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"code":    "UNAUTHORIZED",
					"message": "No identity in context",
				})
			}

			// Get object from path/context
			object := cfg.ObjectFunc(c)
			if object == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"code":    "BAD_REQUEST",
					"message": "No object specified",
				})
			}

			// Check permission
			allowed, err := cfg.Keto.CheckPermission(
				c.Request().Context(),
				identity.ID,
				cfg.Relation,
				cfg.Namespace,
				object,
			)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"code":    "INTERNAL_ERROR",
					"message": "Permission check failed",
				})
			}

			if !allowed {
				return c.JSON(http.StatusForbidden, map[string]string{
					"code":    "FORBIDDEN",
					"message": "Permission denied",
				})
			}

			return next(c)
		}
	}
}

// RequirePermission creates an authz middleware for a specific permission
func RequirePermission(keto PermissionChecker, namespace, relation string, objectFunc func(c echo.Context) string) echo.MiddlewareFunc {
	return Authz(&AuthzConfig{
		Keto:       keto,
		Namespace:  namespace,
		Relation:   relation,
		ObjectFunc: objectFunc,
	})
}

// ObjectFromParam creates an object function that reads from path param
func ObjectFromParam(param string) func(c echo.Context) string {
	return func(c echo.Context) string {
		return c.Param(param)
	}
}

// ObjectFromQuery creates an object function that reads from query param
func ObjectFromQuery(param string) func(c echo.Context) string {
	return func(c echo.Context) string {
		return c.QueryParam(param)
	}
}
