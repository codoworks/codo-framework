package migrations

import (
	"github.com/codoworks/codo-framework/core/db/migrations"
)

// CreateGroupsTable returns the migration for creating the groups table.
func CreateGroupsTable() *migrations.Migration {
	return migrations.NewMigration("20251219000001", "create_groups_table").
		WithUpSQL(`
			CREATE TABLE groups (
				id VARCHAR(36) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				description TEXT,
				color VARCHAR(7),
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				deleted_at TIMESTAMP NULL
			);
			CREATE INDEX idx_groups_deleted_at ON groups(deleted_at);
			CREATE INDEX idx_groups_name ON groups(name);
		`).
		WithDownSQL(`
			DROP INDEX IF EXISTS idx_groups_name;
			DROP INDEX IF EXISTS idx_groups_deleted_at;
			DROP TABLE IF EXISTS groups;
		`)
}
