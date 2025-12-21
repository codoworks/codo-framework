package forms

import (
	"time"

	"github.com/codoworks/codo-framework/examples/models"
)

// CreateGroupRequest is the form for creating a new group
type CreateGroupRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"omitempty,hexcolor|max=7"`
}

// ToModel creates a Group model from the form
func (f *CreateGroupRequest) ToModel() *models.Group {
	return &models.Group{
		Name:        f.Name,
		Description: f.Description,
		Color:       f.Color,
	}
}

// UpdateGroupRequest is the form for updating an existing group
type UpdateGroupRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	Color       *string `json:"color" validate:"omitempty,hexcolor|max=7"`
}

// ApplyTo applies the form data to an existing group
func (f *UpdateGroupRequest) ApplyTo(group *models.Group) {
	if f.Name != nil {
		group.Name = *f.Name
	}
	if f.Description != nil {
		group.Description = *f.Description
	}
	if f.Color != nil {
		group.Color = *f.Color
	}
}

// GroupResponse is the response form for a group
type GroupResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Color        string    `json:"color,omitempty"`
	ContactCount *int64    `json:"contact_count,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FromModel populates the response from a group model
func (f *GroupResponse) FromModel(group *models.Group) *GroupResponse {
	f.ID = group.ID
	f.Name = group.Name
	f.Description = group.Description
	f.Color = group.Color
	f.CreatedAt = group.CreatedAt
	f.UpdatedAt = group.UpdatedAt
	return f
}

// WithCount sets the contact count on the response
func (f *GroupResponse) WithCount(count int64) *GroupResponse {
	f.ContactCount = &count
	return f
}

// NewGroupResponse creates a GroupResponse from a group model
func NewGroupResponse(group *models.Group) *GroupResponse {
	resp := &GroupResponse{}
	return resp.FromModel(group)
}

// NewGroupListResponse creates a list of GroupResponses
func NewGroupListResponse(groups []*models.Group) []*GroupResponse {
	responses := make([]*GroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = NewGroupResponse(group)
	}
	return responses
}

// GroupSummaryResponse is a compact response for group listings
type GroupSummaryResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// FromModel populates the summary from a group model
func (f *GroupSummaryResponse) FromModel(group *models.Group) *GroupSummaryResponse {
	f.ID = group.ID
	f.Name = group.Name
	f.Color = group.Color
	return f
}
