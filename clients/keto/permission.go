package keto

// Permission represents a permission check request
type Permission struct {
	Subject   string
	Relation  string
	Namespace string
	Object    string
}

// NewPermission creates a new permission
func NewPermission(subject, relation, namespace, object string) *Permission {
	return &Permission{
		Subject:   subject,
		Relation:  relation,
		Namespace: namespace,
		Object:    object,
	}
}

// Common relations
const (
	RelationOwner  = "owner"
	RelationEditor = "editor"
	RelationViewer = "viewer"
	RelationMember = "member"
	RelationAdmin  = "admin"
)

// Common namespaces
const (
	NamespaceOrganizations = "organizations"
	NamespaceProjects      = "projects"
	NamespaceResources     = "resources"
)
