package forms

import (
	"testing"
	"time"

	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/examples/models"
)

func TestCreateGroupRequest_ToModel(t *testing.T) {
	tests := []struct {
		name string
		form CreateGroupRequest
	}{
		{
			name: "full form",
			form: CreateGroupRequest{
				Name:        "Friends",
				Description: "Close friends",
				Color:       "#FF5733",
			},
		},
		{
			name: "minimal form",
			form: CreateGroupRequest{
				Name: "Work",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tt.form.ToModel()

			if model.Name != tt.form.Name {
				t.Errorf("Name = %v, want %v", model.Name, tt.form.Name)
			}
			if model.Description != tt.form.Description {
				t.Errorf("Description = %v, want %v", model.Description, tt.form.Description)
			}
			if model.Color != tt.form.Color {
				t.Errorf("Color = %v, want %v", model.Color, tt.form.Color)
			}
		})
	}
}

func TestUpdateGroupRequest_ApplyTo(t *testing.T) {
	name := "Updated Friends"
	description := "Very close friends"
	color := "#00FF00"

	group := &models.Group{
		Name:        "Friends",
		Description: "Close friends",
		Color:       "#FF5733",
	}

	form := UpdateGroupRequest{
		Name:        &name,
		Description: &description,
		Color:       &color,
	}

	form.ApplyTo(group)

	if group.Name != name {
		t.Errorf("Name = %v, want %v", group.Name, name)
	}
	if group.Description != description {
		t.Errorf("Description = %v, want %v", group.Description, description)
	}
	if group.Color != color {
		t.Errorf("Color = %v, want %v", group.Color, color)
	}
}

func TestUpdateGroupRequest_ApplyTo_PartialUpdate(t *testing.T) {
	name := "Updated Friends"

	group := &models.Group{
		Name:        "Friends",
		Description: "Close friends",
		Color:       "#FF5733",
	}

	form := UpdateGroupRequest{
		Name: &name,
		// other fields are nil
	}

	form.ApplyTo(group)

	if group.Name != name {
		t.Errorf("Name = %v, want %v", group.Name, name)
	}
	if group.Description != "Close friends" {
		t.Errorf("Description should be unchanged, got %v", group.Description)
	}
	if group.Color != "#FF5733" {
		t.Errorf("Color should be unchanged, got %v", group.Color)
	}
}

func TestGroupResponse_FromModel(t *testing.T) {
	now := time.Now()

	group := &models.Group{
		Model: db.Model{
			ID:        "group-123",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:        "Friends",
		Description: "Close friends",
		Color:       "#FF5733",
	}

	resp := &GroupResponse{}
	resp.FromModel(group)

	if resp.ID != group.ID {
		t.Errorf("ID = %v, want %v", resp.ID, group.ID)
	}
	if resp.Name != group.Name {
		t.Errorf("Name = %v, want %v", resp.Name, group.Name)
	}
	if resp.Description != group.Description {
		t.Errorf("Description = %v, want %v", resp.Description, group.Description)
	}
	if resp.Color != group.Color {
		t.Errorf("Color = %v, want %v", resp.Color, group.Color)
	}
	if !resp.CreatedAt.Equal(group.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", resp.CreatedAt, group.CreatedAt)
	}
	if !resp.UpdatedAt.Equal(group.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", resp.UpdatedAt, group.UpdatedAt)
	}
	if resp.ContactCount != nil {
		t.Errorf("ContactCount should be nil initially")
	}
}

func TestGroupResponse_WithCount(t *testing.T) {
	group := &models.Group{
		Model: db.Model{ID: "group-123"},
		Name:  "Friends",
	}

	resp := NewGroupResponse(group).WithCount(42)

	if resp.ContactCount == nil {
		t.Error("ContactCount should not be nil")
		return
	}
	if *resp.ContactCount != 42 {
		t.Errorf("ContactCount = %v, want %v", *resp.ContactCount, 42)
	}
}

func TestNewGroupResponse(t *testing.T) {
	group := &models.Group{
		Model: db.Model{ID: "group-123"},
		Name:  "Friends",
	}

	resp := NewGroupResponse(group)
	if resp.ID != group.ID {
		t.Errorf("ID = %v, want %v", resp.ID, group.ID)
	}
	if resp.Name != group.Name {
		t.Errorf("Name = %v, want %v", resp.Name, group.Name)
	}
}

func TestNewGroupListResponse(t *testing.T) {
	groups := []*models.Group{
		{Model: db.Model{ID: "1"}, Name: "Friends"},
		{Model: db.Model{ID: "2"}, Name: "Work"},
	}

	responses := NewGroupListResponse(groups)

	if len(responses) != len(groups) {
		t.Errorf("Length = %v, want %v", len(responses), len(groups))
	}

	for i, resp := range responses {
		if resp.ID != groups[i].ID {
			t.Errorf("responses[%d].ID = %v, want %v", i, resp.ID, groups[i].ID)
		}
	}
}

func TestNewGroupListResponse_Empty(t *testing.T) {
	responses := NewGroupListResponse([]*models.Group{})
	if len(responses) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(responses))
	}
}

func TestGroupSummaryResponse_FromModel(t *testing.T) {
	group := &models.Group{
		Model: db.Model{ID: "group-123"},
		Name:  "Friends",
		Color: "#FF5733",
	}

	resp := &GroupSummaryResponse{}
	resp.FromModel(group)

	if resp.ID != group.ID {
		t.Errorf("ID = %v, want %v", resp.ID, group.ID)
	}
	if resp.Name != group.Name {
		t.Errorf("Name = %v, want %v", resp.Name, group.Name)
	}
	if resp.Color != group.Color {
		t.Errorf("Color = %v, want %v", resp.Color, group.Color)
	}
}
