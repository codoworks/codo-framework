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
			"name": map[string]any{
				"first": "Test",
				"last":  "User",
			},
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

func TestIdentity_FirstName(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"name": map[string]any{
				"first": "John",
				"last":  "Doe",
			},
		},
	}

	assert.Equal(t, "John", identity.FirstName())
}

func TestIdentity_FirstName_Missing(t *testing.T) {
	identity := &Identity{
		ID:     "user-123",
		Traits: map[string]any{},
	}

	assert.Equal(t, "", identity.FirstName())
}

func TestIdentity_FirstName_NoNameTrait(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"email": "test@example.com",
		},
	}

	assert.Equal(t, "", identity.FirstName())
}

func TestIdentity_LastName(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"name": map[string]any{
				"first": "John",
				"last":  "Doe",
			},
		},
	}

	assert.Equal(t, "Doe", identity.LastName())
}

func TestIdentity_LastName_Missing(t *testing.T) {
	identity := &Identity{
		ID:     "user-123",
		Traits: map[string]any{},
	}

	assert.Equal(t, "", identity.LastName())
}

func TestIdentity_Name(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"name": map[string]any{
				"first": "John",
				"last":  "Doe",
			},
		},
	}

	assert.Equal(t, "John Doe", identity.Name())
}

func TestIdentity_Name_FirstOnly(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"name": map[string]any{
				"first": "John",
			},
		},
	}

	assert.Equal(t, "John", identity.Name())
}

func TestIdentity_Name_LastOnly(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"name": map[string]any{
				"last": "Doe",
			},
		},
	}

	assert.Equal(t, "Doe", identity.Name())
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
			"name": map[string]any{
				"first": "Test",
				"last":  "User",
			},
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
	name := traits["name"].(map[string]any)
	assert.Equal(t, "Test", name["first"])
	assert.Equal(t, "User", name["last"])
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

func TestIdentity_SessionID(t *testing.T) {
	identity := &Identity{
		ID:        "user-123",
		SessionID: "session-456",
		Traits: map[string]any{
			"email": "test@example.com",
		},
	}

	assert.Equal(t, "session-456", identity.SessionID)
}

func TestIdentity_MarshalJSON_WithSessionID(t *testing.T) {
	identity := &Identity{
		ID:        "user-123",
		SessionID: "session-456",
		Traits: map[string]any{
			"email": "test@example.com",
		},
	}

	data, err := json.Marshal(identity)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "user-123", result["id"])
	assert.Equal(t, "session-456", result["session_id"])
}

func TestIdentity_MarshalJSON_OmitEmptySessionID(t *testing.T) {
	identity := &Identity{
		ID: "user-123",
		Traits: map[string]any{
			"email": "test@example.com",
		},
	}

	data, err := json.Marshal(identity)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "user-123", result["id"])
	_, hasSessionID := result["session_id"]
	assert.False(t, hasSessionID, "session_id should be omitted when empty")
}

func TestIdentity_UnmarshalJSON_WithSessionID(t *testing.T) {
	data := []byte(`{"id":"user-456","session_id":"session-789","traits":{"email":"other@example.com"}}`)

	var identity Identity
	err := json.Unmarshal(data, &identity)
	require.NoError(t, err)

	assert.Equal(t, "user-456", identity.ID)
	assert.Equal(t, "session-789", identity.SessionID)
	assert.Equal(t, "other@example.com", identity.GetTraitString("email"))
}
