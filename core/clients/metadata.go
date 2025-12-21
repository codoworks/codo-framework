package clients

import "sync"

// ClientRequirement defines whether a client is required or optional
type ClientRequirement string

const (
	// ClientRequired indicates the client must be initialized successfully
	ClientRequired ClientRequirement = "required"
	// ClientOptional indicates the client can fail without stopping the application
	ClientOptional ClientRequirement = "optional"
)

// ClientMetadata holds metadata about a client's requirements
type ClientMetadata struct {
	Name        string
	Requirement ClientRequirement
	FeatureName string // Links to config.Features (e.g., "rabbitmq")
}

var (
	clientMetadata = make(map[string]ClientMetadata)
	metaMu         sync.RWMutex
)

// RegisterMetadata registers metadata for a client
func RegisterMetadata(meta ClientMetadata) {
	metaMu.Lock()
	defer metaMu.Unlock()
	clientMetadata[meta.Name] = meta
}

// GetMetadata retrieves metadata for a client
func GetMetadata(name string) (ClientMetadata, bool) {
	metaMu.RLock()
	defer metaMu.RUnlock()
	meta, exists := clientMetadata[name]
	return meta, exists
}

// IsRequired returns true if the client is required
// Defaults to true for backward compatibility if no metadata is registered
func IsRequired(name string) bool {
	meta, exists := GetMetadata(name)
	if !exists {
		return true // Default to required for backward compatibility
	}
	return meta.Requirement == ClientRequired
}

// ResetMetadata resets all client metadata. For testing only.
func ResetMetadata() {
	metaMu.Lock()
	defer metaMu.Unlock()
	clientMetadata = make(map[string]ClientMetadata)
}
