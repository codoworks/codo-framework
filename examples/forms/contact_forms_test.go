package forms

import (
	"testing"
	"time"

	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/examples/models"
)

func TestCreateContactRequest_ToModel(t *testing.T) {
	groupID := "group-123"

	tests := []struct {
		name string
		form CreateContactRequest
	}{
		{
			name: "full form",
			form: CreateContactRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Phone:     "+1234567890",
				Notes:     "Test notes",
				GroupID:   &groupID,
			},
		},
		{
			name: "minimal form",
			form: CreateContactRequest{
				FirstName: "Jane",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tt.form.ToModel()

			if model.FirstName != tt.form.FirstName {
				t.Errorf("FirstName = %v, want %v", model.FirstName, tt.form.FirstName)
			}
			if model.LastName != tt.form.LastName {
				t.Errorf("LastName = %v, want %v", model.LastName, tt.form.LastName)
			}
			if model.Email != tt.form.Email {
				t.Errorf("Email = %v, want %v", model.Email, tt.form.Email)
			}
			if model.Phone != tt.form.Phone {
				t.Errorf("Phone = %v, want %v", model.Phone, tt.form.Phone)
			}
			if model.Notes != tt.form.Notes {
				t.Errorf("Notes = %v, want %v", model.Notes, tt.form.Notes)
			}

			if tt.form.GroupID != nil {
				if model.GroupID == nil || *model.GroupID != *tt.form.GroupID {
					t.Errorf("GroupID mismatch")
				}
			}
		})
	}
}

func TestUpdateContactRequest_ApplyTo(t *testing.T) {
	firstName := "Jane"
	lastName := "Smith"
	email := "jane@example.com"
	phone := "+0987654321"
	notes := "Updated notes"
	groupID := "new-group"
	emptyGroupID := ""

	contact := &models.Contact{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "+1234567890",
		Notes:     "Original notes",
	}

	form := UpdateContactRequest{
		FirstName: &firstName,
		LastName:  &lastName,
		Email:     &email,
		Phone:     &phone,
		Notes:     &notes,
		GroupID:   &groupID,
	}

	form.ApplyTo(contact)

	if contact.FirstName != firstName {
		t.Errorf("FirstName = %v, want %v", contact.FirstName, firstName)
	}
	if contact.LastName != lastName {
		t.Errorf("LastName = %v, want %v", contact.LastName, lastName)
	}
	if contact.Email != email {
		t.Errorf("Email = %v, want %v", contact.Email, email)
	}
	if contact.Phone != phone {
		t.Errorf("Phone = %v, want %v", contact.Phone, phone)
	}
	if contact.Notes != notes {
		t.Errorf("Notes = %v, want %v", contact.Notes, notes)
	}
	if contact.GroupID == nil || *contact.GroupID != groupID {
		t.Errorf("GroupID mismatch")
	}

	// Test clearing group with empty string
	emptyForm := UpdateContactRequest{
		GroupID: &emptyGroupID,
	}
	emptyForm.ApplyTo(contact)

	if contact.GroupID != nil {
		t.Errorf("GroupID should be nil after setting empty string")
	}
}

func TestUpdateContactRequest_ApplyTo_PartialUpdate(t *testing.T) {
	firstName := "Jane"

	contact := &models.Contact{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	form := UpdateContactRequest{
		FirstName: &firstName,
		// other fields are nil
	}

	form.ApplyTo(contact)

	if contact.FirstName != firstName {
		t.Errorf("FirstName = %v, want %v", contact.FirstName, firstName)
	}
	if contact.LastName != "Doe" {
		t.Errorf("LastName should be unchanged, got %v", contact.LastName)
	}
	if contact.Email != "john@example.com" {
		t.Errorf("Email should be unchanged, got %v", contact.Email)
	}
}

func TestContactResponse_FromModel(t *testing.T) {
	groupID := "group-123"
	now := time.Now()

	contact := &models.Contact{
		Model: db.Model{
			ID:        "contact-123",
			CreatedAt: now,
			UpdatedAt: now,
		},
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "+1234567890",
		Notes:     "Some notes",
		GroupID:   &groupID,
	}

	resp := &ContactResponse{}
	resp.FromModel(contact)

	if resp.ID != contact.ID {
		t.Errorf("ID = %v, want %v", resp.ID, contact.ID)
	}
	if resp.FirstName != contact.FirstName {
		t.Errorf("FirstName = %v, want %v", resp.FirstName, contact.FirstName)
	}
	if resp.LastName != contact.LastName {
		t.Errorf("LastName = %v, want %v", resp.LastName, contact.LastName)
	}
	if resp.FullName != "John Doe" {
		t.Errorf("FullName = %v, want %v", resp.FullName, "John Doe")
	}
	if resp.Email != contact.Email {
		t.Errorf("Email = %v, want %v", resp.Email, contact.Email)
	}
	if resp.Phone != contact.Phone {
		t.Errorf("Phone = %v, want %v", resp.Phone, contact.Phone)
	}
	if resp.Notes != contact.Notes {
		t.Errorf("Notes = %v, want %v", resp.Notes, contact.Notes)
	}
	if resp.GroupID == nil || *resp.GroupID != groupID {
		t.Errorf("GroupID mismatch")
	}
	if !resp.CreatedAt.Equal(contact.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", resp.CreatedAt, contact.CreatedAt)
	}
	if !resp.UpdatedAt.Equal(contact.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", resp.UpdatedAt, contact.UpdatedAt)
	}
}

func TestNewContactResponse(t *testing.T) {
	contact := &models.Contact{
		Model: db.Model{
			ID: "contact-123",
		},
		FirstName: "John",
	}

	resp := NewContactResponse(contact)
	if resp.ID != contact.ID {
		t.Errorf("ID = %v, want %v", resp.ID, contact.ID)
	}
}

func TestNewContactListResponse(t *testing.T) {
	contacts := []*models.Contact{
		{Model: db.Model{ID: "1"}, FirstName: "John"},
		{Model: db.Model{ID: "2"}, FirstName: "Jane"},
	}

	responses := NewContactListResponse(contacts)

	if len(responses) != len(contacts) {
		t.Errorf("Length = %v, want %v", len(responses), len(contacts))
	}

	for i, resp := range responses {
		if resp.ID != contacts[i].ID {
			t.Errorf("responses[%d].ID = %v, want %v", i, resp.ID, contacts[i].ID)
		}
	}
}

func TestNewContactListResponse_Empty(t *testing.T) {
	responses := NewContactListResponse([]*models.Contact{})
	if len(responses) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(responses))
	}
}
