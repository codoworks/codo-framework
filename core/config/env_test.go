package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvVarRegistry_Register(t *testing.T) {
	registry := NewEnvVarRegistry()

	t.Run("registers valid descriptor", func(t *testing.T) {
		err := registry.Register(EnvVarDescriptor{
			Name:     "TEST_VAR_1",
			Type:     EnvTypeString,
			Required: true,
			Group:    "test",
		})
		assert.NoError(t, err)
	})

	t.Run("rejects empty name", func(t *testing.T) {
		err := registry.Register(EnvVarDescriptor{
			Name: "",
			Type: EnvTypeString,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("rejects duplicate registration", func(t *testing.T) {
		registry := NewEnvVarRegistry()
		err := registry.Register(EnvVarDescriptor{
			Name: "DUPLICATE_VAR",
			Type: EnvTypeString,
		})
		require.NoError(t, err)

		err = registry.Register(EnvVarDescriptor{
			Name: "DUPLICATE_VAR",
			Type: EnvTypeString,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestEnvVarRegistry_RegisterMany(t *testing.T) {
	registry := NewEnvVarRegistry()

	descs := []EnvVarDescriptor{
		{Name: "VAR_A", Type: EnvTypeString},
		{Name: "VAR_B", Type: EnvTypeInt},
		{Name: "VAR_C", Type: EnvTypeBool},
	}

	err := registry.RegisterMany(descs)
	assert.NoError(t, err)

	all := registry.All()
	assert.Len(t, all, 3)
}

func TestEnvVarRegistry_Resolve_String(t *testing.T) {
	t.Run("resolves required string", func(t *testing.T) {
		os.Setenv("RESOLVE_STRING_TEST", "hello world")
		defer os.Unsetenv("RESOLVE_STRING_TEST")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "RESOLVE_STRING_TEST",
			Type:     EnvTypeString,
			Required: true,
		}))

		err := r.Resolve()
		assert.NoError(t, err)

		val, err := r.GetString("RESOLVE_STRING_TEST")
		assert.NoError(t, err)
		assert.Equal(t, "hello world", val)
	})

	t.Run("fails on missing required string", func(t *testing.T) {
		os.Unsetenv("MISSING_REQUIRED_STRING")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "MISSING_REQUIRED_STRING",
			Type:     EnvTypeString,
			Required: true,
		}))

		err := r.Resolve()
		assert.Error(t, err)
		assert.True(t, r.HasErrors())
	})

	t.Run("uses default for optional string", func(t *testing.T) {
		os.Unsetenv("OPTIONAL_STRING_DEFAULT")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "OPTIONAL_STRING_DEFAULT",
			Type:     EnvTypeString,
			Required: false,
			Default:  "default value",
		}))

		err := r.Resolve()
		assert.NoError(t, err)

		val, err := r.GetString("OPTIONAL_STRING_DEFAULT")
		assert.NoError(t, err)
		assert.Equal(t, "default value", val)
	})
}

func TestEnvVarRegistry_Resolve_Int(t *testing.T) {
	t.Run("resolves valid int", func(t *testing.T) {
		os.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "TEST_INT",
			Type:     EnvTypeInt,
			Required: true,
		}))

		err := r.Resolve()
		assert.NoError(t, err)

		val, err := r.GetInt("TEST_INT")
		assert.NoError(t, err)
		assert.Equal(t, 42, val)
	})

	t.Run("fails on invalid int", func(t *testing.T) {
		os.Setenv("INVALID_INT", "not-a-number")
		defer os.Unsetenv("INVALID_INT")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "INVALID_INT",
			Type:     EnvTypeInt,
			Required: true,
		}))

		err := r.Resolve()
		assert.Error(t, err)
	})
}

func TestEnvVarRegistry_Resolve_Bool(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"TRUE", true},
		{"FALSE", false},
	}

	for _, tc := range testCases {
		t.Run("parses "+tc.value, func(t *testing.T) {
			os.Setenv("TEST_BOOL", tc.value)
			defer os.Unsetenv("TEST_BOOL")

			r := NewEnvVarRegistry()
			require.NoError(t, r.Register(EnvVarDescriptor{
				Name:     "TEST_BOOL",
				Type:     EnvTypeBool,
				Required: true,
			}))

			err := r.Resolve()
			assert.NoError(t, err)

			val, err := r.GetBool("TEST_BOOL")
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, val)
		})
	}
}

func TestEnvVarRegistry_Resolve_Duration(t *testing.T) {
	t.Run("resolves valid duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION", "30s")
		defer os.Unsetenv("TEST_DURATION")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "TEST_DURATION",
			Type:     EnvTypeDuration,
			Required: true,
		}))

		err := r.Resolve()
		assert.NoError(t, err)

		val, err := r.GetDuration("TEST_DURATION")
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, val)
	})

	t.Run("resolves complex duration", func(t *testing.T) {
		os.Setenv("TEST_DURATION_COMPLEX", "1h30m45s")
		defer os.Unsetenv("TEST_DURATION_COMPLEX")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "TEST_DURATION_COMPLEX",
			Type:     EnvTypeDuration,
			Required: true,
		}))

		err := r.Resolve()
		assert.NoError(t, err)

		val, err := r.GetDuration("TEST_DURATION_COMPLEX")
		assert.NoError(t, err)
		expected := 1*time.Hour + 30*time.Minute + 45*time.Second
		assert.Equal(t, expected, val)
	})
}

func TestEnvVarRegistry_Resolve_URL(t *testing.T) {
	t.Run("resolves valid URL", func(t *testing.T) {
		os.Setenv("TEST_URL", "https://api.example.com:8080/v1")
		defer os.Unsetenv("TEST_URL")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "TEST_URL",
			Type:     EnvTypeURL,
			Required: true,
		}))

		err := r.Resolve()
		assert.NoError(t, err)

		val, err := r.GetURL("TEST_URL")
		assert.NoError(t, err)
		assert.Equal(t, "https", val.Scheme)
		assert.Equal(t, "api.example.com:8080", val.Host)
		assert.Equal(t, "/v1", val.Path)
	})

	t.Run("fails on URL without scheme", func(t *testing.T) {
		os.Setenv("BAD_URL", "api.example.com/v1")
		defer os.Unsetenv("BAD_URL")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "BAD_URL",
			Type:     EnvTypeURL,
			Required: true,
		}))

		err := r.Resolve()
		assert.Error(t, err)
	})
}

func TestEnvVarRegistry_CustomValidator(t *testing.T) {
	t.Run("passes with valid value", func(t *testing.T) {
		os.Setenv("VALIDATED_VAR", "hello")
		defer os.Unsetenv("VALIDATED_VAR")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "VALIDATED_VAR",
			Type:     EnvTypeString,
			Required: true,
			Validator: func(value any) error {
				if s, ok := value.(string); ok && len(s) < 3 {
					return assert.AnError
				}
				return nil
			},
		}))

		err := r.Resolve()
		assert.NoError(t, err)
	})

	t.Run("fails with invalid value", func(t *testing.T) {
		os.Setenv("VALIDATED_VAR_FAIL", "hi")
		defer os.Unsetenv("VALIDATED_VAR_FAIL")

		r := NewEnvVarRegistry()
		require.NoError(t, r.Register(EnvVarDescriptor{
			Name:     "VALIDATED_VAR_FAIL",
			Type:     EnvTypeString,
			Required: true,
			Validator: func(value any) error {
				if s, ok := value.(string); ok && len(s) < 3 {
					return assert.AnError
				}
				return nil
			},
		}))

		err := r.Resolve()
		assert.Error(t, err)
	})
}

func TestEnvVarRegistry_Groups(t *testing.T) {
	os.Setenv("STRIPE_KEY", "sk_test_123")
	os.Setenv("STRIPE_SECRET", "whsec_456")
	os.Setenv("OTHER_VAR", "other")
	defer func() {
		os.Unsetenv("STRIPE_KEY")
		os.Unsetenv("STRIPE_SECRET")
		os.Unsetenv("OTHER_VAR")
	}()

	r := NewEnvVarRegistry()
	require.NoError(t, r.RegisterMany([]EnvVarDescriptor{
		{Name: "STRIPE_KEY", Type: EnvTypeString, Required: true, Group: "stripe"},
		{Name: "STRIPE_SECRET", Type: EnvTypeString, Required: true, Group: "stripe"},
		{Name: "OTHER_VAR", Type: EnvTypeString, Required: true, Group: "other"},
	}))

	err := r.Resolve()
	require.NoError(t, err)

	stripeVars := r.GetGroup("stripe")
	assert.Len(t, stripeVars, 2)
	assert.Equal(t, "sk_test_123", stripeVars["STRIPE_KEY"].Value)
	assert.Equal(t, "whsec_456", stripeVars["STRIPE_SECRET"].Value)

	otherVars := r.GetGroup("other")
	assert.Len(t, otherVars, 1)

	emptyVars := r.GetGroup("nonexistent")
	assert.Len(t, emptyVars, 0)
}

func TestEnvVarRegistry_MaskedValues(t *testing.T) {
	os.Setenv("PUBLIC_VAR", "public value")
	os.Setenv("SECRET_VAR", "super-secret-key")
	defer func() {
		os.Unsetenv("PUBLIC_VAR")
		os.Unsetenv("SECRET_VAR")
	}()

	r := NewEnvVarRegistry()
	require.NoError(t, r.RegisterMany([]EnvVarDescriptor{
		{Name: "PUBLIC_VAR", Type: EnvTypeString, Required: true, Sensitive: false},
		{Name: "SECRET_VAR", Type: EnvTypeString, Required: true, Sensitive: true},
	}))

	err := r.Resolve()
	require.NoError(t, err)

	masked := r.MaskedValues()
	assert.Equal(t, "public value", masked["PUBLIC_VAR"])
	assert.Equal(t, "***MASKED***", masked["SECRET_VAR"])
}

func TestEnvVarValue_TypedAccessors(t *testing.T) {
	os.Setenv("TYPED_STRING", "hello")
	os.Setenv("TYPED_INT", "42")
	os.Setenv("TYPED_BOOL", "true")
	os.Setenv("TYPED_DURATION", "5m")
	os.Setenv("TYPED_FLOAT", "3.14")
	defer func() {
		os.Unsetenv("TYPED_STRING")
		os.Unsetenv("TYPED_INT")
		os.Unsetenv("TYPED_BOOL")
		os.Unsetenv("TYPED_DURATION")
		os.Unsetenv("TYPED_FLOAT")
	}()

	r := NewEnvVarRegistry()
	require.NoError(t, r.RegisterMany([]EnvVarDescriptor{
		{Name: "TYPED_STRING", Type: EnvTypeString, Required: true},
		{Name: "TYPED_INT", Type: EnvTypeInt, Required: true},
		{Name: "TYPED_BOOL", Type: EnvTypeBool, Required: true},
		{Name: "TYPED_DURATION", Type: EnvTypeDuration, Required: true},
		{Name: "TYPED_FLOAT", Type: EnvTypeFloat, Required: true},
	}))

	err := r.Resolve()
	require.NoError(t, err)

	val, _ := r.Get("TYPED_STRING")
	assert.Equal(t, "hello", val.String())

	val, _ = r.Get("TYPED_INT")
	assert.Equal(t, 42, val.Int())

	val, _ = r.Get("TYPED_BOOL")
	assert.Equal(t, true, val.Bool())

	val, _ = r.Get("TYPED_DURATION")
	assert.Equal(t, 5*time.Minute, val.Duration())

	val, _ = r.Get("TYPED_FLOAT")
	assert.Equal(t, 3.14, val.Float())
}

func TestEnvValidationErrors(t *testing.T) {
	t.Run("collects multiple errors", func(t *testing.T) {
		os.Unsetenv("MISSING_1")
		os.Unsetenv("MISSING_2")
		os.Setenv("INVALID_INT_VAL", "not-int")
		defer os.Unsetenv("INVALID_INT_VAL")

		r := NewEnvVarRegistry()
		require.NoError(t, r.RegisterMany([]EnvVarDescriptor{
			{Name: "MISSING_1", Type: EnvTypeString, Required: true},
			{Name: "MISSING_2", Type: EnvTypeString, Required: true},
			{Name: "INVALID_INT_VAL", Type: EnvTypeInt, Required: true},
		}))

		err := r.Resolve()
		assert.Error(t, err)

		errs := r.ValidationErrors()
		assert.Len(t, errs, 3)
	})

	t.Run("formats error message", func(t *testing.T) {
		var errs EnvValidationErrors
		errs.Add("VAR_A", "is required")
		errs.Add("VAR_B", "invalid format")

		msg := errs.Error()
		assert.Contains(t, msg, "VAR_A: is required")
		assert.Contains(t, msg, "VAR_B: invalid format")
	})

	t.Run("converts to framework error", func(t *testing.T) {
		var errs EnvValidationErrors
		errs.Add("TEST_VAR", "required variable not set")

		fwkErr := errs.ToFrameworkError()
		assert.NotNil(t, fwkErr)
		assert.Equal(t, "VALIDATION_ERROR", fwkErr.Code)
		assert.Contains(t, fwkErr.Message, "Environment variable validation failed")
	})
}

func TestBuiltInValidators(t *testing.T) {
	t.Run("NotEmpty", func(t *testing.T) {
		v := NotEmpty()
		assert.NoError(t, v("hello"))
		assert.Error(t, v(""))
		assert.Error(t, v("   "))
	})

	t.Run("MinLength", func(t *testing.T) {
		v := MinLength(5)
		assert.NoError(t, v("hello"))
		assert.NoError(t, v("hello world"))
		assert.Error(t, v("hi"))
	})

	t.Run("MaxLength", func(t *testing.T) {
		v := MaxLength(5)
		assert.NoError(t, v("hello"))
		assert.NoError(t, v("hi"))
		assert.Error(t, v("hello world"))
	})

	t.Run("OneOf", func(t *testing.T) {
		v := OneOf("dev", "staging", "prod")
		assert.NoError(t, v("dev"))
		assert.NoError(t, v("prod"))
		assert.Error(t, v("test"))
	})

	t.Run("IntRange", func(t *testing.T) {
		v := IntRange(1, 100)
		assert.NoError(t, v(50))
		assert.NoError(t, v(1))
		assert.NoError(t, v(100))
		assert.Error(t, v(0))
		assert.Error(t, v(101))
	})

	t.Run("Port", func(t *testing.T) {
		v := Port()
		assert.NoError(t, v(8080))
		assert.NoError(t, v(1))
		assert.NoError(t, v(65535))
		assert.Error(t, v(0))
		assert.Error(t, v(65536))
	})

	t.Run("DurationMin", func(t *testing.T) {
		v := DurationMin(time.Second)
		assert.NoError(t, v(5 * time.Second))
		assert.NoError(t, v(time.Second))
		assert.Error(t, v(500 * time.Millisecond))
	})

	t.Run("Chain", func(t *testing.T) {
		v := Chain(NotEmpty(), MinLength(3), MaxLength(10))
		assert.NoError(t, v("hello"))
		assert.Error(t, v(""))      // fails NotEmpty
		assert.Error(t, v("hi"))    // fails MinLength
		assert.Error(t, v("hello world")) // fails MaxLength
	})
}

func TestEnvVarRegistry_Reset(t *testing.T) {
	r := NewEnvVarRegistry()

	os.Setenv("RESET_TEST", "value")
	defer os.Unsetenv("RESET_TEST")

	require.NoError(t, r.Register(EnvVarDescriptor{
		Name:     "RESET_TEST",
		Type:     EnvTypeString,
		Required: true,
	}))

	require.NoError(t, r.Resolve())
	assert.True(t, r.IsResolved())

	r.Reset()
	assert.False(t, r.IsResolved())
	assert.Len(t, r.All(), 0)
}
