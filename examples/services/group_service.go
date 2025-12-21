package services

import (
	"context"

	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/examples/models"
)

// GroupService handles business logic for contact groups
type GroupService struct {
	repo        *db.Repository[*models.Group]
	contactRepo *db.Repository[*models.Contact]
}

// NewGroupService creates a new GroupService
func NewGroupService(client *db.Client) *GroupService {
	return &GroupService{
		repo:        db.NewRepository[*models.Group](client),
		contactRepo: db.NewRepository[*models.Contact](client),
	}
}

// Create creates a new group
func (s *GroupService) Create(ctx context.Context, group *models.Group) error {
	return s.repo.Create(ctx, group)
}

// Update updates an existing group
func (s *GroupService) Update(ctx context.Context, group *models.Group) error {
	return s.repo.Update(ctx, group)
}

// Delete soft-deletes a group
// Note: This does not delete contacts in the group, they become ungrouped
func (s *GroupService) Delete(ctx context.Context, group *models.Group) error {
	// Unassign all contacts from this group first
	_, err := s.contactRepo.UpdateWhere(ctx,
		map[string]any{"group_id": nil},
		db.Where("group_id = ?", group.ID),
	)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, group)
}

// FindByID retrieves a group by ID
func (s *GroupService) FindByID(ctx context.Context, id string) (*models.Group, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return record.Model(), nil
}

// FindAll retrieves all groups
func (s *GroupService) FindAll(ctx context.Context, opts ...db.QueryOption) ([]*models.Group, error) {
	records, err := s.repo.FindAll(ctx, opts...)
	if err != nil {
		return nil, err
	}

	groups := make([]*models.Group, len(records))
	for i, r := range records {
		groups[i] = r.Model()
	}
	return groups, nil
}

// Count returns the total number of groups
func (s *GroupService) Count(ctx context.Context, opts ...db.QueryOption) (int64, error) {
	return s.repo.Count(ctx, opts...)
}

// FindByName finds a group by name
func (s *GroupService) FindByName(ctx context.Context, name string) (*models.Group, error) {
	record, err := s.repo.FindOne(ctx, db.Where("name = ?", name))
	if err != nil {
		return nil, err
	}
	return record.Model(), nil
}

// GetContactCount returns the number of contacts in a group
func (s *GroupService) GetContactCount(ctx context.Context, groupID string) (int64, error) {
	return s.contactRepo.Count(ctx, db.Where("group_id = ?", groupID))
}

// GroupWithCount represents a group with its contact count
type GroupWithCount struct {
	Group *models.Group
	Count int64
}

// FindAllWithCounts retrieves all groups with their contact counts
func (s *GroupService) FindAllWithCounts(ctx context.Context, opts ...db.QueryOption) ([]GroupWithCount, error) {
	groups, err := s.FindAll(ctx, opts...)
	if err != nil {
		return nil, err
	}

	results := make([]GroupWithCount, len(groups))
	for i, group := range groups {
		count, err := s.GetContactCount(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		results[i] = GroupWithCount{
			Group: group,
			Count: count,
		}
	}

	return results, nil
}
