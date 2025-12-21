package models

import (
	"testing"
	"time"

	"github.com/codoworks/codo-framework/core/db"
)

func TestContact_TableName(t *testing.T) {
	c := &Contact{}
	if got := c.TableName(); got != "contacts" {
		t.Errorf("TableName() = %v, want %v", got, "contacts")
	}
}

func TestContact_FullName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      string
	}{
		{"both names", "John", "Doe", "John Doe"},
		{"first name only", "John", "", "John"},
		{"last name only", "", "Doe", "Doe"},
		{"no names", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Contact{
				FirstName: tt.firstName,
				LastName:  tt.lastName,
			}
			if got := c.FullName(); got != tt.want {
				t.Errorf("FullName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContact_HasEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"with email", "test@example.com", true},
		{"empty email", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Contact{Email: tt.email}
			if got := c.HasEmail(); got != tt.want {
				t.Errorf("HasEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContact_HasPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{"with phone", "+1234567890", true},
		{"empty phone", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Contact{Phone: tt.phone}
			if got := c.HasPhone(); got != tt.want {
				t.Errorf("HasPhone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContact_BelongsToGroup(t *testing.T) {
	groupID := "group-123"
	emptyGroupID := ""

	tests := []struct {
		name    string
		groupID *string
		want    bool
	}{
		{"with group", &groupID, true},
		{"nil group", nil, false},
		{"empty group", &emptyGroupID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Contact{GroupID: tt.groupID}
			if got := c.BelongsToGroup(); got != tt.want {
				t.Errorf("BelongsToGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContact_EmbeddedModel(t *testing.T) {
	c := &Contact{
		Model: db.Model{
			ID:        "contact-123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FirstName: "John",
		LastName:  "Doe",
	}

	if c.GetID() != "contact-123" {
		t.Errorf("GetID() = %v, want %v", c.GetID(), "contact-123")
	}

	if c.PrimaryKey() != "id" {
		t.Errorf("PrimaryKey() = %v, want %v", c.PrimaryKey(), "id")
	}

	if c.IsNew() {
		t.Error("Contact with ID should not be new")
	}

	if !c.IsPersisted() {
		t.Error("Contact with ID and CreatedAt should be persisted")
	}
}

func TestContact_Modeler(t *testing.T) {
	var _ db.Modeler = &Contact{}
}
