package models

import (
	"testing"
	"time"

	"github.com/codoworks/codo-framework/core/db"
)

func TestGroup_TableName(t *testing.T) {
	g := &Group{}
	if got := g.TableName(); got != "groups" {
		t.Errorf("TableName() = %v, want %v", got, "groups")
	}
}

func TestGroup_HasDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        bool
	}{
		{"with description", "Some description", true},
		{"empty description", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Group{Description: tt.description}
			if got := g.HasDescription(); got != tt.want {
				t.Errorf("HasDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroup_HasColor(t *testing.T) {
	tests := []struct {
		name  string
		color string
		want  bool
	}{
		{"with color", "#FF5733", true},
		{"empty color", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Group{Color: tt.color}
			if got := g.HasColor(); got != tt.want {
				t.Errorf("HasColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroup_EmbeddedModel(t *testing.T) {
	g := &Group{
		Model: db.Model{
			ID:        "group-123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Friends",
		Description: "Close friends",
		Color:       "#00FF00",
	}

	if g.GetID() != "group-123" {
		t.Errorf("GetID() = %v, want %v", g.GetID(), "group-123")
	}

	if g.PrimaryKey() != "id" {
		t.Errorf("PrimaryKey() = %v, want %v", g.PrimaryKey(), "id")
	}

	if g.IsNew() {
		t.Error("Group with ID should not be new")
	}

	if !g.IsPersisted() {
		t.Error("Group with ID and CreatedAt should be persisted")
	}
}

func TestGroup_Modeler(t *testing.T) {
	var _ db.Modeler = &Group{}
}
