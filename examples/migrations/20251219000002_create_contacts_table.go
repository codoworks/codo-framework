package migrations

import (
	"github.com/codoworks/codo-framework/core/db/migrations"
)

// CreateContactsTable returns the migration for creating the contacts table.
func CreateContactsTable() *migrations.Migration {
	return migrations.NewMigration("20251219000002", "create_contacts_table").
		WithUpSQL(`
			CREATE TABLE contacts (
				id VARCHAR(36) PRIMARY KEY,
				first_name VARCHAR(255) NOT NULL,
				last_name VARCHAR(255),
				email VARCHAR(255),
				phone VARCHAR(50),
				notes TEXT,
				group_id VARCHAR(36),
				created_at TIMESTAMP NOT NULL,
				updated_at TIMESTAMP NOT NULL,
				deleted_at TIMESTAMP NULL,
				FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL
			);
			CREATE INDEX idx_contacts_deleted_at ON contacts(deleted_at);
			CREATE INDEX idx_contacts_group_id ON contacts(group_id);
			CREATE INDEX idx_contacts_email ON contacts(email);
			CREATE INDEX idx_contacts_name ON contacts(first_name, last_name);
		`).
		WithDownSQL(`
			DROP INDEX IF EXISTS idx_contacts_name;
			DROP INDEX IF EXISTS idx_contacts_email;
			DROP INDEX IF EXISTS idx_contacts_group_id;
			DROP INDEX IF EXISTS idx_contacts_deleted_at;
			DROP TABLE IF EXISTS contacts;
		`)
}
