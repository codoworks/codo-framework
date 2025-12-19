package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Context extends echo.Context with additional helpers
type Context struct {
	echo.Context
}

// BindAndValidate binds the request body and validates it
func (c *Context) BindAndValidate(form any) error {
	if err := c.Bind(form); err != nil {
		return &BindError{Cause: err}
	}
	if err := c.Validate(form); err != nil {
		return err
	}
	return nil
}

// ParamUUID extracts a UUID path parameter
func (c *Context) ParamUUID(name string) (string, error) {
	val := c.Param(name)
	if val == "" {
		return "", &ParamError{Param: name, Message: "required"}
	}

	// Validate UUID format
	if _, err := uuid.Parse(val); err != nil {
		return "", &ParamError{Param: name, Message: "must be a valid UUID"}
	}

	return val, nil
}

// QueryInt extracts an integer query parameter with default
func (c *Context) QueryInt(name string, defaultVal int) int {
	val := c.QueryParam(name)
	if val == "" {
		return defaultVal
	}

	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// QueryInt64 extracts an int64 query parameter with default
func (c *Context) QueryInt64(name string, defaultVal int64) int64 {
	val := c.QueryParam(name)
	if val == "" {
		return defaultVal
	}

	result, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultVal
	}
	return result
}

// QueryBool extracts a boolean query parameter with default
func (c *Context) QueryBool(name string, defaultVal bool) bool {
	val := c.QueryParam(name)
	if val == "" {
		return defaultVal
	}

	result, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// Success sends a 200 OK response
func (c *Context) Success(payload any) error {
	return c.JSON(http.StatusOK, Success(payload))
}

// Created sends a 201 Created response
func (c *Context) Created(payload any) error {
	return c.JSON(http.StatusCreated, Created(payload))
}

// NoContent sends a 204 No Content response
func (c *Context) NoContent() error {
	return c.Context.NoContent(http.StatusNoContent)
}

// SendError sends an error response
func (c *Context) SendError(err error) error {
	resp := ErrorResponse(err)
	return c.JSON(resp.HTTPStatus, resp)
}

// GetRequestID returns the request ID from context
func (c *Context) GetRequestID() string {
	if id := c.Request().Header.Get("X-Request-ID"); id != "" {
		return id
	}
	if id := c.Response().Header().Get("X-Request-ID"); id != "" {
		return id
	}
	return ""
}

// RealIP returns the client's real IP address
func (c *Context) RealIP() string {
	return c.Context.RealIP()
}

// BindError represents a binding error
type BindError struct {
	Cause error
}

func (e *BindError) Error() string {
	return fmt.Sprintf("binding error: %v", e.Cause)
}

// Unwrap returns the underlying cause
func (e *BindError) Unwrap() error {
	return e.Cause
}

// ParamError represents a parameter error
type ParamError struct {
	Param   string
	Message string
}

func (e *ParamError) Error() string {
	return fmt.Sprintf("parameter %s: %s", e.Param, e.Message)
}
