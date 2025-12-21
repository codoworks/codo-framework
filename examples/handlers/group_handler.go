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

// GroupHandler handles HTTP requests for contact groups
type GroupHandler struct {
	service        *services.GroupService
	contactService *services.ContactService
}

// NewGroupHandler creates a new GroupHandler
func NewGroupHandler(dbClient *db.Client) *GroupHandler {
	return &GroupHandler{
		service:        services.NewGroupService(dbClient),
		contactService: services.NewContactService(dbClient),
	}
}

// Prefix returns the URL prefix for group routes
func (h *GroupHandler) Prefix() string {
	return "/api/v1/groups"
}

// Scope returns the router scope (Protected - requires auth)
func (h *GroupHandler) Scope() http.RouterScope {
	return http.ScopeProtected
}

// Middlewares returns handler-specific middlewares
func (h *GroupHandler) Middlewares() []echo.MiddlewareFunc {
	return nil
}

// Initialize performs any required initialization
func (h *GroupHandler) Initialize() error {
	return nil
}

// Routes registers the handler's routes
func (h *GroupHandler) Routes(g *echo.Group) {
	g.GET("", http.WrapHandler(h.List))
	g.POST("", http.WrapHandler(h.Create))
	g.GET("/:id", http.WrapHandler(h.Get))
	g.PUT("/:id", http.WrapHandler(h.Update))
	g.DELETE("/:id", http.WrapHandler(h.Delete))
	g.GET("/:id/contacts", http.WrapHandler(h.ListContacts))
}

// List returns a paginated list of groups with contact counts
func (h *GroupHandler) List(c *http.Context) error {
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	// Get total count
	total, err := h.service.Count(c.Request().Context())
	if err != nil {
		return c.SendError(err)
	}

	// Add pagination
	limit, offset := perPage, (page-1)*perPage
	opts := []db.QueryOption{
		db.Limit(limit),
		db.Offset(offset),
		db.OrderByAsc("name"),
	}

	// Fetch groups with counts
	groupsWithCounts, err := h.service.FindAllWithCounts(c.Request().Context(), opts...)
	if err != nil {
		return c.SendError(err)
	}

	// Build response
	items := make([]*forms.GroupResponse, len(groupsWithCounts))
	for i, gwc := range groupsWithCounts {
		items[i] = forms.NewGroupResponse(gwc.Group).WithCount(gwc.Count)
	}

	response := coreforms.NewListResponse(items, total, page, perPage)

	return c.Success(response)
}

// Create creates a new group
func (h *GroupHandler) Create(c *http.Context) error {
	var form forms.CreateGroupRequest
	if err := c.BindAndValidate(&form); err != nil {
		return c.SendError(err)
	}

	group := form.ToModel()
	if err := h.service.Create(c.Request().Context(), group); err != nil {
		return c.SendError(err)
	}

	return c.Created(forms.NewGroupResponse(group))
}

// Get retrieves a single group by ID
func (h *GroupHandler) Get(c *http.Context) error {
	id, err := c.ParamUUID("id")
	if err != nil {
		return c.SendError(err)
	}

	group, err := h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusNotFound, http.NotFound("Group not found"))
		}
		return c.SendError(err)
	}

	// Get contact count
	count, err := h.service.GetContactCount(c.Request().Context(), id)
	if err != nil {
		return c.SendError(err)
	}

	return c.Success(forms.NewGroupResponse(group).WithCount(count))
}

// Update updates an existing group
func (h *GroupHandler) Update(c *http.Context) error {
	id, err := c.ParamUUID("id")
	if err != nil {
		return c.SendError(err)
	}

	// Find existing group
	group, err := h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusNotFound, http.NotFound("Group not found"))
		}
		return c.SendError(err)
	}

	// Bind and validate form
	var form forms.UpdateGroupRequest
	if err := c.BindAndValidate(&form); err != nil {
		return c.SendError(err)
	}

	// Apply updates
	form.ApplyTo(group)

	if err := h.service.Update(c.Request().Context(), group); err != nil {
		return c.SendError(err)
	}

	// Get contact count
	count, err := h.service.GetContactCount(c.Request().Context(), id)
	if err != nil {
		return c.SendError(err)
	}

	return c.Success(forms.NewGroupResponse(group).WithCount(count))
}

// Delete soft-deletes a group and unassigns its contacts
func (h *GroupHandler) Delete(c *http.Context) error {
	id, err := c.ParamUUID("id")
	if err != nil {
		return c.SendError(err)
	}

	group, err := h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusNotFound, http.NotFound("Group not found"))
		}
		return c.SendError(err)
	}

	if err := h.service.Delete(c.Request().Context(), group); err != nil {
		return c.SendError(err)
	}

	return c.NoContent()
}

// ListContacts returns contacts belonging to a group
func (h *GroupHandler) ListContacts(c *http.Context) error {
	id, err := c.ParamUUID("id")
	if err != nil {
		return c.SendError(err)
	}

	// Verify group exists
	_, err = h.service.FindByID(c.Request().Context(), id)
	if err != nil {
		if err == db.ErrNotFound {
			return c.JSON(nethttp.StatusNotFound, http.NotFound("Group not found"))
		}
		return c.SendError(err)
	}

	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 20)

	// Get contacts in group
	whereGroup := db.Where("group_id = ?", id)

	total, err := h.contactService.Count(c.Request().Context(), whereGroup)
	if err != nil {
		return c.SendError(err)
	}

	limit, offset := perPage, (page-1)*perPage
	contacts, err := h.contactService.FindByGroup(c.Request().Context(), id,
		db.Limit(limit), db.Offset(offset), db.OrderByDesc("created_at"))
	if err != nil {
		return c.SendError(err)
	}

	items := forms.NewContactListResponse(contacts)
	response := coreforms.NewListResponse(items, total, page, perPage)

	return c.Success(response)
}

