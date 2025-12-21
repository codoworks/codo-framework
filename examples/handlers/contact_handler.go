package handlers

import (
	nethttp "net/http"

	"github.com/labstack/echo/v4"

	"github.com/codoworks/codo-framework/core/db"
	coreforms "github.com/codoworks/codo-framework/core/forms"
	"github.com/codoworks/codo-framework/core/http"
	"github.com/codoworks/codo-framework/examples/forms"
	"github.com/codoworks/codo-framework/examples/services"
)

// ContactHandler handles HTTP requests for contacts
type ContactHandler struct {
	service *services.ContactService
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(dbClient *db.Client) *ContactHandler {
	return &ContactHandler{
		service: services.NewContactService(dbClient),
	}
}

// Prefix returns the URL prefix for contact routes
func (h *ContactHandler) Prefix() string {
	return "/api/v1/contacts"
}

// Scope returns the router scope (Protected - requires auth)
func (h *ContactHandler) Scope() http.RouterScope {
	return http.ScopeProtected
}

// Middlewares returns handler-specific middlewares
func (h *ContactHandler) Middlewares() []echo.MiddlewareFunc {
	return nil
}

// Initialize performs any required initialization
func (h *ContactHandler) Initialize() error {
	return nil
}

// Routes registers the handler's routes
func (h *ContactHandler) Routes(g *echo.Group) {
	g.GET("", http.WrapHandler(h.List))
	g.POST("", http.WrapHandler(h.Create))
	g.GET("/search", http.WrapHandler(h.Search))
	g.GET("/:id", http.WrapHandler(h.Get))
	g.PUT("/:id", http.WrapHandler(h.Update))
	g.DELETE("/:id", http.WrapHandler(h.Delete))
	g.POST("/:id/move", http.WrapHandler(h.Move))
}

// List returns a paginated list of contacts
func (h *ContactHandler) List(c *http.Context) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)
	groupID := c.QueryParam("group_id")

	// Build query options
	var opts []db.QueryOption
	if groupID != "" {
		opts = append(opts, db.Where("group_id = ?", groupID))
	}

	// Get total count
	total, err := h.service.Count(c.Request().Context(), opts...)
	if err != nil {
		return c.SendError(err)
	}

	// Add pagination
	limit, offset := perPage, (page-1)*perPage
	opts = append(opts, db.Limit(limit), db.Offset(offset), db.OrderByDesc("created_at"))

	// Fetch contacts
	contacts, err := h.service.FindAll(c.Request().Context(), opts...)
	if err != nil {
		return c.SendError(err)
	}

	// Build response
	items := forms.NewContactListResponse(contacts)
	response := coreforms.NewListResponse(items, total, page, perPage)

	return c.Success(response)
}

// Create creates a new contact
func (h *ContactHandler) Create(c *http.Context) error {
	var form forms.CreateContactRequest
	if err := c.BindAndValidate(&form); err != nil {
		return c.SendError(err)
	}

	contact := form.ToModel()
	if err := h.service.Create(c.Request().Context(), contact); err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusUnprocessableEntity, http.ValidationError([]string{"group_id: group not found"}))
		}
		return c.SendError(err)
	}

	return c.Created(forms.NewContactResponse(contact))
}

// Get retrieves a single contact by ID
func (h *ContactHandler) Get(c *http.Context) error {
	id, err := c.ParamUUID("id")
	if err != nil {
		return c.SendError(err)
	}

	contact, err := h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusNotFound, http.NotFound("Contact not found"))
		}
		return c.SendError(err)
	}

	return c.Success(forms.NewContactResponse(contact))
}

// Update updates an existing contact
func (h *ContactHandler) Update(c *http.Context) error {
	id, err := c.ParamUUID("id")
	if err != nil {
		return c.SendError(err)
	}

	// Find existing contact
	contact, err := h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusNotFound, http.NotFound("Contact not found"))
		}
		return c.SendError(err)
	}

	// Bind and validate form
	var form forms.UpdateContactRequest
	if err := c.BindAndValidate(&form); err != nil {
		return c.SendError(err)
	}

	// Apply updates
	form.ApplyTo(contact)

	if err := h.service.Update(c.Request().Context(), contact); err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusUnprocessableEntity, http.ValidationError([]string{"group_id: group not found"}))
		}
		return c.SendError(err)
	}

	return c.Success(forms.NewContactResponse(contact))
}

// Delete soft-deletes a contact
func (h *ContactHandler) Delete(c *http.Context) error {
	id, err := c.ParamUUID("id")
	if err != nil {
		return c.SendError(err)
	}

	contact, err := h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusNotFound, http.NotFound("Contact not found"))
		}
		return c.SendError(err)
	}

	if err := h.service.Delete(c.Request().Context(), contact); err != nil {
		return c.SendError(err)
	}

	return c.NoContent()
}

// Search searches contacts by name, email, or phone
func (h *ContactHandler) Search(c *http.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(nethttp.StatusBadRequest, http.BadRequest("Search query required"))
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	// Get total count for search
	searchPattern := "%" + query + "%"
	whereClause := db.Where(
		"first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR phone LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)

	total, err := h.service.Count(c.Request().Context(), whereClause)
	if err != nil {
		return c.SendError(err)
	}

	// Search with pagination
	limit, offset := perPage, (page-1)*perPage
	contacts, err := h.service.Search(c.Request().Context(), query,
		db.Limit(limit), db.Offset(offset), db.OrderByDesc("created_at"))
	if err != nil {
		return c.SendError(err)
	}

	items := forms.NewContactListResponse(contacts)
	response := coreforms.NewListResponse(items, total, page, perPage)

	return c.Success(response)
}

// Move moves a contact to a different group
func (h *ContactHandler) Move(c *http.Context) error {
	id, err := c.ParamUUID("id")
	if err != nil {
		return c.SendError(err)
	}

	var form forms.MoveContactRequest
	if err := c.BindAndValidate(&form); err != nil {
		return c.SendError(err)
	}

	if err := h.service.MoveToGroup(c.Request().Context(), id, form.GroupID); err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusNotFound, http.NotFound("Contact or group not found"))
		}
		return c.SendError(err)
	}

	// Fetch updated contact
	contact, err := h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		return c.SendError(err)
	}

	return c.Success(forms.NewContactResponse(contact))
}

