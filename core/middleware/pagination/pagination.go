package pagination

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/codoworks/codo-framework/clients/logger"
	"github.com/codoworks/codo-framework/core/clients"
	"github.com/codoworks/codo-framework/core/config"
	"github.com/codoworks/codo-framework/core/middleware"
	"github.com/codoworks/codo-framework/core/pagination"
)

func init() {
	middleware.RegisterMiddleware(&PaginationMiddleware{
		BaseMiddleware: middleware.NewBaseMiddleware(
			"pagination",
			"middleware.pagination",
			middleware.PriorityPagination,
			middleware.RouterAll,
		),
	})
}

// PaginationMiddleware extracts and validates pagination parameters from requests
type PaginationMiddleware struct {
	middleware.BaseMiddleware

	defaultPageSize int
	maxPageSize     int
	defaultType     pagination.Type
	logDetails      bool
	paramNames      config.PaginationParamNames
	logger          *logrus.Logger
}

// Enabled checks if pagination middleware is enabled in config.
// Returns false by default - consumers must explicitly enable.
func (m *PaginationMiddleware) Enabled(cfg any) bool {
	if cfg == nil {
		return false // DISABLED BY DEFAULT
	}

	paginationCfg, ok := cfg.(*config.PaginationMiddlewareConfig)
	if !ok {
		return false
	}

	return paginationCfg.Enabled
}

// Configure initializes the middleware with configuration
func (m *PaginationMiddleware) Configure(cfg any) error {
	// Set defaults
	m.defaultPageSize = 20
	m.maxPageSize = 100
	m.defaultType = pagination.TypeOffset
	m.logDetails = false
	m.paramNames = config.PaginationParamNames{
		Page:      "page",
		PerPage:   "per_page",
		Cursor:    "cursor",
		Direction: "direction",
	}

	// Get logger client for optional logging (don't fail if not available)
	if loggerClient, err := clients.GetTyped[*logger.Logger]("logger"); err == nil {
		m.logger = loggerClient.GetLogger()
	}

	// Override with config if provided
	paginationCfg, ok := cfg.(*config.PaginationMiddlewareConfig)
	if !ok {
		return nil
	}

	if paginationCfg.DefaultPageSize > 0 {
		m.defaultPageSize = paginationCfg.DefaultPageSize
	}
	if paginationCfg.MaxPageSize > 0 {
		m.maxPageSize = paginationCfg.MaxPageSize
	}
	if paginationCfg.DefaultType == "cursor" {
		m.defaultType = pagination.TypeCursor
	}
	m.logDetails = paginationCfg.LogDetails

	if paginationCfg.ParamNames.Page != "" {
		m.paramNames.Page = paginationCfg.ParamNames.Page
	}
	if paginationCfg.ParamNames.PerPage != "" {
		m.paramNames.PerPage = paginationCfg.ParamNames.PerPage
	}
	if paginationCfg.ParamNames.Cursor != "" {
		m.paramNames.Cursor = paginationCfg.ParamNames.Cursor
	}
	if paginationCfg.ParamNames.Direction != "" {
		m.paramNames.Direction = paginationCfg.ParamNames.Direction
	}

	return nil
}

// Handler returns the pagination middleware function
func (m *PaginationMiddleware) Handler() echo.MiddlewareFunc {
	// Capture config values in closure
	defaultPageSize := m.defaultPageSize
	maxPageSize := m.maxPageSize
	defaultType := m.defaultType
	logDetails := m.logDetails
	paramNames := m.paramNames
	log := m.logger

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only process GET requests - pagination makes sense for list endpoints
			if c.Request().Method != http.MethodGet {
				return next(c)
			}

			params := extractParams(c, defaultPageSize, maxPageSize, defaultType, paramNames)
			pagination.Set(c, params)

			// Optional logging
			if logDetails && log != nil {
				if params.IsOffset() {
					log.WithFields(logrus.Fields{
						"type":     "offset",
						"page":     params.Page,
						"per_page": params.PerPage,
						"offset":   params.Offset,
					}).Debug("pagination")
				} else {
					log.WithFields(logrus.Fields{
						"type":      "cursor",
						"cursor":    params.Cursor,
						"direction": params.Direction,
						"per_page":  params.PerPage,
					}).Debug("pagination")
				}
			}

			return next(c)
		}
	}
}

// extractParams extracts and validates pagination parameters from the request
func extractParams(
	c echo.Context,
	defaultPageSize, maxPageSize int,
	defaultType pagination.Type,
	paramNames config.PaginationParamNames,
) *pagination.Params {
	params := &pagination.Params{
		MaxPerPage: maxPageSize,
		Type:       defaultType,
	}

	// Extract raw values
	params.RawPage = c.QueryParam(paramNames.Page)
	params.RawPerPage = c.QueryParam(paramNames.PerPage)
	params.RawCursor = c.QueryParam(paramNames.Cursor)

	// Determine pagination type based on presence of cursor
	if params.RawCursor != "" {
		params.Type = pagination.TypeCursor
		params.Cursor = params.RawCursor

		// Parse direction
		dir := c.QueryParam(paramNames.Direction)
		if dir == "prev" {
			params.Direction = pagination.DirectionPrev
		} else {
			params.Direction = pagination.DirectionNext
		}
	}

	// Parse per_page
	if params.RawPerPage != "" {
		if perPage, err := strconv.Atoi(params.RawPerPage); err == nil {
			params.PerPage = perPage
		}
	}
	if params.PerPage <= 0 {
		params.PerPage = defaultPageSize
	}
	if params.PerPage > maxPageSize {
		params.PerPage = maxPageSize
	}

	// Parse page (only relevant for offset-based)
	if params.Type == pagination.TypeOffset {
		if params.RawPage != "" {
			if page, err := strconv.Atoi(params.RawPage); err == nil {
				params.Page = page
			}
		}
		if params.Page <= 0 {
			params.Page = 1
		}
		params.Offset = (params.Page - 1) * params.PerPage
	}

	return params
}
