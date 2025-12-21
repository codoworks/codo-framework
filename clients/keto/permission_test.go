package keto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPermission(t *testing.T) {
	perm := NewPermission("user-123", "viewer", "documents", "doc-456")

	assert.Equal(t, "user-123", perm.Subject)
	assert.Equal(t, "viewer", perm.Relation)
	assert.Equal(t, "documents", perm.Namespace)
	assert.Equal(t, "doc-456", perm.Object)
}

func TestPermission_Fields(t *testing.T) {
	perm := &Permission{
		Subject:   "user-abc",
		Relation:  "editor",
		Namespace: "projects",
		Object:    "proj-xyz",
	}

	assert.Equal(t, "user-abc", perm.Subject)
	assert.Equal(t, "editor", perm.Relation)
	assert.Equal(t, "projects", perm.Namespace)
	assert.Equal(t, "proj-xyz", perm.Object)
}

func TestRelationConstants(t *testing.T) {
	assert.Equal(t, "owner", RelationOwner)
	assert.Equal(t, "editor", RelationEditor)
	assert.Equal(t, "viewer", RelationViewer)
	assert.Equal(t, "member", RelationMember)
	assert.Equal(t, "admin", RelationAdmin)
}

func TestNamespaceConstants(t *testing.T) {
	assert.Equal(t, "organizations", NamespaceOrganizations)
	assert.Equal(t, "projects", NamespaceProjects)
	assert.Equal(t, "resources", NamespaceResources)
}

func TestNewPermission_WithConstants(t *testing.T) {
	perm := NewPermission("user-123", RelationViewer, NamespaceProjects, "proj-456")

	assert.Equal(t, "user-123", perm.Subject)
	assert.Equal(t, "viewer", perm.Relation)
	assert.Equal(t, "projects", perm.Namespace)
	assert.Equal(t, "proj-456", perm.Object)
}
