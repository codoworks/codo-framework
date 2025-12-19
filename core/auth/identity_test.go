package auth

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentity_GetTrait(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}

	val, ok := identity.GetTrait("email")
	assert.True(t, ok)
	assert.Equal(t, "test@example.com", val)
}

func TestIdentity_GetTrait_Missing(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"email": "test@example.com",
		},
	}

	val, ok := identity.GetTrait("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestIdentity_GetTrait_NilTraits(t *testing.T) {
	identity := &Identity{
		ID:     "user-123",
		Traits: nil,
	}

	val, ok := identity.GetTrait("email")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestIdentity_GetTraitString(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"email": "test@example.com",
		},
	}

	val := identity.GetTraitString("email")
	assert.Equal(t, "test@example.com", val)
}

func TestIdentity_GetTraitString_NotString(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"count": 42,
		},
	}

	val := identity.GetTraitString("count")
	assert.Equal(t, "", val)
}

func TestIdentity_GetTraitString_Missing(t *testing.T) {
	identity := &Identity{
		ID:     "user-123",
		Traits: map[string]any{},
	}

	val := identity.GetTraitString("nonexistent")
	assert.Equal(t, "", val)
}

func TestIdentity_GetTraitBool(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"verified": true,
		},
	}

	val := identity.GetTraitBool("verified")
	assert.True(t, val)
}

func TestIdentity_GetTraitBool_False(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"verified": false,
		},
	}

	val := identity.GetTraitBool("verified")
	assert.False(t, val)
}

func TestIdentity_GetTraitBool_NotBool(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"verified": "yes",
		},
	}

	val := identity.GetTraitBool("verified")
	assert.False(t, val)
}

func TestIdentity_GetTraitBool_Missing(t *testing.T) {
	identity := &Identity{
		ID:     "user-123",
		Traits: map[string]any{},
	}

	val := identity.GetTraitBool("nonexistent")
	assert.False(t, val)
}

func TestIdentity_Email(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"email": "test@example.com",
		},
	}

	assert.Equal(t, "test@example.com", identity.Email())
}

func TestIdentity_Email_Missing(t *testing.T) {
	identity := &Identity{
		ID:     "user-123",
		Traits: map[string]any{},
	}

	assert.Equal(t, "", identity.Email())
}

func TestIdentity_Name(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"name": "Test User",
		},
	}

	assert.Equal(t, "Test User", identity.Name())
}

func TestIdentity_Name_Missing(t *testing.T) {
	identity := &Identity{
		ID:     "user-123",
		Traits: map[string]any{},
	}

	assert.Equal(t, "", identity.Name())
}

func TestIdentity_MarshalJSON(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"email": "test@example.com",
			"name":  "Test User",
		},
	}

	data, err := json.Marshal(identity)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "user-123", result["id"])
	traits := result["traits"].(map[string]any)
	assert.Equal(t, "test@example.com", traits["email"])
	assert.Equal(t, "Test User", traits["name"])
}

func TestIdentity_UnmarshalJSON(t *testing.T) {
	data := []byte(`{"id":"user-456","traits":{"email":"other@example.com","verified":true}}`)

	var identity Identity
	err := json.Unmarshal(data, &identity)
	require.NoError(t, err)

	assert.Equal(t, "user-456", identity.ID)
	assert.Equal(t, "other@example.com", identity.GetTraitString("email"))
	assert.True(t, identity.GetTraitBool("verified"))
}

func TestIdentity_UnmarshalJSON_EmptyTraits(t *testing.T) {
	data := []byte(`{"id":"user-789"}`)

	var identity Identity
	err := json.Unmarshal(data, &identity)
	require.NoError(t, err)

	assert.Equal(t, "user-789", identity.ID)
	assert.Nil(t, identity.Traits)
}

func TestIdentity_UnmarshalJSON_Invalid(t *testing.T) {
	data := []byte(`{invalid json}`)

	var identity Identity
	err := json.Unmarshal(data, &identity)
	assert.Error(t, err)
}
