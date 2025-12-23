package auth

import (
	"encoding/json"
)

// Identity represents an authenticated user
type Identity struct {
	ID     string         `json:"id"`
	Traits map[string]any `json:"traits"`
}

// GetTrait retrieves a trait value by key
func (i *Identity) GetTrait(key string) (any, bool) {
	if i.Traits == nil {
		return nil, false
	}
	val, ok := i.Traits[key]
	return val, ok
}

// GetTraitString retrieves a string trait value
func (i *Identity) GetTraitString(key string) string {
	val, ok := i.GetTrait(key)
	if !ok {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetTraitBool retrieves a boolean trait value
func (i *Identity) GetTraitBool(key string) bool {
	val, ok := i.GetTrait(key)
	if !ok {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// Email returns the email trait if present
func (i *Identity) Email() string {
	return i.GetTraitString("email")
}

// getNameMap extracts the nested name object from traits
func (i *Identity) getNameMap() map[string]any {
	val, ok := i.GetTrait("name")
	if !ok {
		return nil
	}
	if m, ok := val.(map[string]any); ok {
		return m
	}
	return nil
}

// FirstName returns the first name from nested name.first trait
func (i *Identity) FirstName() string {
	m := i.getNameMap()
	if m == nil {
		return ""
	}
	if s, ok := m["first"].(string); ok {
		return s
	}
	return ""
}

// LastName returns the last name from nested name.last trait
func (i *Identity) LastName() string {
	m := i.getNameMap()
	if m == nil {
		return ""
	}
	if s, ok := m["last"].(string); ok {
		return s
	}
	return ""
}

// Name returns the full name (first + last) from nested name trait
func (i *Identity) Name() string {
	first := i.FirstName()
	last := i.LastName()
	if first == "" && last == "" {
		return ""
	}
	if first == "" {
		return last
	}
	if last == "" {
		return first
	}
	return first + " " + last
}

// MarshalJSON implements json.Marshaler
func (i *Identity) MarshalJSON() ([]byte, error) {
	type Alias Identity
	return json.Marshal((*Alias)(i))
}

// UnmarshalJSON implements json.Unmarshaler
func (i *Identity) UnmarshalJSON(data []byte) error {
	type Alias Identity
	aux := (*Alias)(i)
	return json.Unmarshal(data, aux)
}
