package seeds

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/codoworks/codo-framework/core/db/seeds"
)

// DefaultGroups returns a seed that creates default contact groups.
func DefaultGroups() *seeds.Seed {
	return seeds.NewSeed("default_groups", func(ctx context.Context, db *sqlx.DB) error {
		groups := []struct {
			Name        string
			Description string
			Color       string
		}{
			{
				Name:        "Friends",
				Description: "Personal friends and social contacts",
				Color:       "#3B82F6", // Blue
			},
			{
				Name:        "Family",
				Description: "Family members and relatives",
				Color:       "#EF4444", // Red
			},
			{
				Name:        "Work",
				Description: "Professional and business contacts",
				Color:       "#10B981", // Green
			},
		}

		now := time.Now()

		for _, g := range groups {
			_, err := db.ExecContext(ctx, `
				INSERT INTO groups (id, name, description, color, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (id) DO NOTHING
			`, uuid.NewString(), g.Name, g.Description, g.Color, now, now)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
