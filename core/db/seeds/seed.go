package seeds

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Seed represents a database seeder
type Seed struct {
	Name string
	Run  func(ctx context.Context, db *sqlx.DB) error
}

// NewSeed creates a new seed
func NewSeed(name string, run func(ctx context.Context, db *sqlx.DB) error) *Seed {
	return &Seed{
		Name: name,
		Run:  run,
	}
}

// Seeder manages and runs seeds
type Seeder struct {
	db    *sqlx.DB
	seeds []*Seed
}

// NewSeeder creates a new seeder
func NewSeeder(db *sqlx.DB) *Seeder {
	return &Seeder{
		db:    db,
		seeds: make([]*Seed, 0),
	}
}

// Add adds seeds to the seeder
func (s *Seeder) Add(seeds ...*Seed) *Seeder {
	s.seeds = append(s.seeds, seeds...)
	return s
}

// Seeds returns the registered seeds
func (s *Seeder) Seeds() []*Seed {
	return s.seeds
}

// Run executes all registered seeds
func (s *Seeder) Run(ctx context.Context) error {
	for _, seed := range s.seeds {
		if err := seed.Run(ctx, s.db); err != nil {
			return fmt.Errorf("seed %s failed: %w", seed.Name, err)
		}
	}
	return nil
}

// RunOne executes a specific seed by name
func (s *Seeder) RunOne(ctx context.Context, name string) error {
	for _, seed := range s.seeds {
		if seed.Name == name {
			return seed.Run(ctx, s.db)
		}
	}
	return fmt.Errorf("seed %s not found", name)
}

// RunByNames executes seeds by their names
func (s *Seeder) RunByNames(ctx context.Context, names ...string) error {
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	for _, seed := range s.seeds {
		if nameSet[seed.Name] {
			if err := seed.Run(ctx, s.db); err != nil {
				return fmt.Errorf("seed %s failed: %w", seed.Name, err)
			}
		}
	}
	return nil
}

// Clear truncates the given tables
func (s *Seeder) Clear(ctx context.Context, tables ...string) error {
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to clear table %s: %w", table, err)
		}
	}
	return nil
}

// Refresh clears tables and re-runs all seeds
func (s *Seeder) Refresh(ctx context.Context, tables ...string) error {
	if err := s.Clear(ctx, tables...); err != nil {
		return err
	}
	return s.Run(ctx)
}

// Count returns the number of registered seeds
func (s *Seeder) Count() int {
	return len(s.seeds)
}

// HasSeed checks if a seed with the given name exists
func (s *Seeder) HasSeed(name string) bool {
	for _, seed := range s.seeds {
		if seed.Name == name {
			return true
		}
	}
	return false
}
