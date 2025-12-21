package models

import (
	"github.com/codoworks/codo-framework/core/db"
)

// Contact represents a contact in the phonebook
type Contact struct {
	db.Model
	FirstName string  `db:"first_name"`
	LastName  string  `db:"last_name"`
	Email     string  `db:"email"`
	Phone     string  `db:"phone"`
	Notes     string  `db:"notes"`
	GroupID   *string `db:"group_id"`
}

// TableName returns the table name for contacts
func (c *Contact) TableName() string {
	return "contacts"
}

// FullName returns the contact's full name
func (c *Contact) FullName() string {
	if c.LastName == "" {
		return c.FirstName
	}
	if c.FirstName == "" {
		return c.LastName
	}
	return c.FirstName + " " + c.LastName
}

// HasEmail returns true if the contact has an email
func (c *Contact) HasEmail() bool {
	return c.Email != ""
}

// HasPhone returns true if the contact has a phone number
func (c *Contact) HasPhone() bool {
	return c.Phone != ""
}

// BelongsToGroup returns true if the contact belongs to a group
func (c *Contact) BelongsToGroup() bool {
	return c.GroupID != nil && *c.GroupID != ""
}
