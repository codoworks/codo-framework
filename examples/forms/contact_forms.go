package forms

import (
	"time"

	"github.com/codoworks/codo-framework/examples/models"
)

// CreateContactRequest is the form for creating a new contact
type CreateContactRequest struct {
	FirstName string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string  `json:"last_name" validate:"max=100"`
	Email     string  `json:"email" validate:"omitempty,email,max=255"`
	Phone     string  `json:"phone" validate:"max=50"`
	Notes     string  `json:"notes" validate:"max=1000"`
	GroupID   *string `json:"group_id" validate:"omitempty,uuid"`
}

// ToModel creates a Contact model from the form
func (f *CreateContactRequest) ToModel() *models.Contact {
	return &models.Contact{
		FirstName: f.FirstName,
		LastName:  f.LastName,
		Email:     f.Email,
		Phone:     f.Phone,
		Notes:     f.Notes,
		GroupID:   f.GroupID,
	}
}

// UpdateContactRequest is the form for updating an existing contact
type UpdateContactRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,max=100"`
	Email     *string `json:"email" validate:"omitempty,email,max=255"`
	Phone     *string `json:"phone" validate:"omitempty,max=50"`
	Notes     *string `json:"notes" validate:"omitempty,max=1000"`
	GroupID   *string `json:"group_id" validate:"omitempty,uuid"`
}

// ApplyTo applies the form data to an existing contact
func (f *UpdateContactRequest) ApplyTo(contact *models.Contact) {
	if f.FirstName != nil {
		contact.FirstName = *f.FirstName
	}
	if f.LastName != nil {
		contact.LastName = *f.LastName
	}
	if f.Email != nil {
		contact.Email = *f.Email
	}
	if f.Phone != nil {
		contact.Phone = *f.Phone
	}
	if f.Notes != nil {
		contact.Notes = *f.Notes
	}
	if f.GroupID != nil {
		if *f.GroupID == "" {
			contact.GroupID = nil
		} else {
			contact.GroupID = f.GroupID
		}
	}
}

// ContactResponse is the response form for a contact
type ContactResponse struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	GroupID   *string   `json:"group_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromModel populates the response from a contact model
func (f *ContactResponse) FromModel(contact *models.Contact) *ContactResponse {
	f.ID = contact.ID
	f.FirstName = contact.FirstName
	f.LastName = contact.LastName
	f.FullName = contact.FullName()
	f.Email = contact.Email
	f.Phone = contact.Phone
	f.Notes = contact.Notes
	f.GroupID = contact.GroupID
	f.CreatedAt = contact.CreatedAt
	f.UpdatedAt = contact.UpdatedAt
	return f
}

// NewContactResponse creates a ContactResponse from a contact model
func NewContactResponse(contact *models.Contact) *ContactResponse {
	resp := &ContactResponse{}
	return resp.FromModel(contact)
}

// NewContactListResponse creates a list of ContactResponses
func NewContactListResponse(contacts []*models.Contact) []*ContactResponse {
	responses := make([]*ContactResponse, len(contacts))
	for i, contact := range contacts {
		responses[i] = NewContactResponse(contact)
	}
	return responses
}

// MoveContactRequest is the form for moving a contact to a group
type MoveContactRequest struct {
	GroupID *string `json:"group_id" validate:"omitempty,uuid"`
}

// BatchMoveRequest is the form for moving multiple contacts to a group
// This is an example of a batch operation that may have partial failures
type BatchMoveRequest struct {
	ContactIDs []string `json:"contact_ids" validate:"required,min=1,max=100,dive,uuid"`
	GroupID    *string  `json:"group_id" validate:"omitempty,uuid"`
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Processed int `json:"processed"` // Number of items successfully processed
	Total     int `json:"total"`     // Total items in request
}
