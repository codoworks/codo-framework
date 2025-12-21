// Package seeds contains database seeds for the contact book application.
package seeds

import (
	"github.com/codoworks/codo-framework/core/db/seeds"
)

// All returns all seeds for the contact book application.
// Seeds are returned in the order they should be executed.
func All() []*seeds.Seed {
	return []*seeds.Seed{
		DefaultGroups(),
	}
}

// AddToSeeder adds all contact book seeds to a seeder.
func AddToSeeder(seeder *seeds.Seeder) {
	seeder.Add(All()...)
}
