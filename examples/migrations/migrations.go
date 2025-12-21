// Package migrations contains database migrations for the contact book application.
package migrations

import (
	"github.com/codoworks/codo-framework/core/db/migrations"
)

// All returns all migrations for the contact book application.
// Migrations are returned in order and should be applied sequentially.
func All() []*migrations.Migration {
	return []*migrations.Migration{
		CreateGroupsTable(),
		CreateContactsTable(),
	}
}

// AddToRunner adds all contact book migrations to a migration runner.
func AddToRunner(runner *migrations.Runner) {
	runner.Add(All()...)
}
