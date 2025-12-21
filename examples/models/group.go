package models

import (
	"github.com/codoworks/codo-framework/core/db"
)

// Group represents a contact group for organizing contacts
type Group struct {
	db.Model
	Name        string `db:"name"`
	Description string `db:"description"`
	Color       string `db:"color"`
}

// TableName returns the table name for groups
func (g *Group) TableName() string {
	return "groups"
}

// HasDescription returns true if the group has a description
func (g *Group) HasDescription() bool {
	return g.Description != ""
}

// HasColor returns true if the group has a color set
func (g *Group) HasColor() bool {
	return g.Color != ""
}
