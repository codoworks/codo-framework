package services

import (
	"context"

	"github.com/codoworks/codo-framework/core/db"
	"github.com/codoworks/codo-framework/examples/models"
)

// ContactService handles business logic for contacts
type ContactService struct {
	repo      *db.Repository[*models.Contact]
	groupRepo *db.Repository[*models.Group]
}

// NewContactService creates a new ContactService
func NewContactService(client *db.Client) *ContactService {
	return &ContactService{
		repo:      db.NewRepository[*models.Contact](client),
		groupRepo: db.NewRepository[*models.Group](client),
	}
}

// Create creates a new contact
func (s *ContactService) Create(ctx context.Context, contact *models.Contact) error {
	// Validate group exists if specified
	if contact.GroupID != nil && *contact.GroupID != "" {
		exists, err := s.groupRepo.Exists(ctx, *contact.GroupID)
		if err != nil {
			return err
		}
		if !exists {
			return db.ErrNotFound
		}
	}

	return s.repo.Create(ctx, contact)
}

// Update updates an existing contact
func (s *ContactService) Update(ctx context.Context, contact *models.Contact) error {
	// Validate group exists if specified
	if contact.GroupID != nil && *contact.GroupID != "" {
		exists, err := s.groupRepo.Exists(ctx, *contact.GroupID)
		if err != nil {
			return err
		}
		if !exists {
			return db.ErrNotFound
		}
	}

	return s.repo.Update(ctx, contact)
}

// Delete soft-deletes a contact
func (s *ContactService) Delete(ctx context.Context, contact *models.Contact) error {
	return s.repo.Delete(ctx, contact)
}

// FindByID retrieves a contact by ID
func (s *ContactService) FindByID(ctx context.Context, id string) (*models.Contact, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return record.Model(), nil
}

// FindAll retrieves all contacts with optional filtering
func (s *ContactService) FindAll(ctx context.Context, opts ...db.QueryOption) ([]*models.Contact, error) {
	records, err := s.repo.FindAll(ctx, opts...)
	if err != nil {
		return nil, err
	}

	contacts := make([]*models.Contact, len(records))
	for i, r := range records {
		contacts[i] = r.Model()
	}
	return contacts, nil
}

// Count returns the total number of contacts
func (s *ContactService) Count(ctx context.Context, opts ...db.QueryOption) (int64, error) {
	return s.repo.Count(ctx, opts...)
}

// FindByGroup retrieves all contacts in a group
func (s *ContactService) FindByGroup(ctx context.Context, groupID string, opts ...db.QueryOption) ([]*models.Contact, error) {
	opts = append([]db.QueryOption{db.Where("group_id = ?", groupID)}, opts...)
	return s.FindAll(ctx, opts...)
}

// FindByEmail finds a contact by email address
func (s *ContactService) FindByEmail(ctx context.Context, email string) (*models.Contact, error) {
	record, err := s.repo.FindOne(ctx, db.Where("email = ?", email))
	if err != nil {
		return nil, err
	}
	return record.Model(), nil
}

// Search searches contacts by name, email, or phone
func (s *ContactService) Search(ctx context.Context, query string, opts ...db.QueryOption) ([]*models.Contact, error) {
	searchPattern := "%" + query + "%"
	whereClause := db.Where(
		"first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR phone LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)
	opts = append([]db.QueryOption{whereClause}, opts...)
	return s.FindAll(ctx, opts...)
}

// MoveToGroup moves a contact to a different group
func (s *ContactService) MoveToGroup(ctx context.Context, contactID string, groupID *string) error {
	record, err := s.repo.FindByID(ctx, contactID)
	if err != nil {
		return err
	}

	contact := record.Model()

	// Validate new group exists if specified
	if groupID != nil && *groupID != "" {
		exists, err := s.groupRepo.Exists(ctx, *groupID)
		if err != nil {
			return err
		}
		if !exists {
			return db.ErrNotFound
		}
	}

	contact.GroupID = groupID
	return s.repo.Update(ctx, contact)
}

// RemoveFromGroup removes a contact from its group
func (s *ContactService) RemoveFromGroup(ctx context.Context, contactID string) error {
	return s.MoveToGroup(ctx, contactID, nil)
}
