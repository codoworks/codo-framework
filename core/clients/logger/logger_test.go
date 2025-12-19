package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/codoworks/codo-framework/core/clients"
	"github.com/stretchr/testify/assert"
)

func TestLogger_ImplementsClient(t *testing.T) {
	var _ clients.Client = (*Logger)(nil)
}

func TestNew(t *testing.T) {
	l := New()

	assert.NotNil(t, l)
	assert.Equal(t, ClientName, l.Name())
	assert.NotNil(t, l.logger)
}

func TestLogger_Name(t *testing.T) {
	l := New()
	assert.Equal(t, "logger", l.Name())
}

func TestLogger_Initialize(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		l := New()

		err := l.Initialize(nil)

		assert.NoError(t, err)
		assert.True(t, l.IsInitialized())
	})

	t.Run("with Config pointer", func(t *testing.T) {
		l := New()
		buf := &bytes.Buffer{}
		cfg := &Config{
			Level:  LevelDebug,
			Format: FormatText,
			Output: buf,
		}

		err := l.Initialize(cfg)

		assert.NoError(t, err)
		assert.Equal(t, LevelDebug, l.GetLevel())
	})

	t.Run("with Config value", func(t *testing.T) {
		l := New()
		cfg := Config{
			Level:  LevelWarn,
			Format: FormatJSON,
		}

		err := l.Initialize(cfg)

		assert.NoError(t, err)
		// logrus returns "warning" not "warn"
		assert.Equal(t, Level("warning"), l.GetLevel())
	})

	t.Run("with map config", func(t *testing.T) {
		l := New()
		cfg := map[string]any{
			"level":  "error",
			"format": "text",
		}

		err := l.Initialize(cfg)

		assert.NoError(t, err)
		assert.Equal(t, LevelError, l.GetLevel())
	})

	t.Run("with invalid config type", func(t *testing.T) {
		l := New()

		err := l.Initialize("invalid")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid config type")
	})

	t.Run("with invalid level uses info", func(t *testing.T) {
		l := New()
		cfg := &Config{
			Level: Level("invalid-level"),
		}

		err := l.Initialize(cfg)

		assert.NoError(t, err)
		assert.Equal(t, LevelInfo, l.GetLevel())
	})
}

func TestLogger_Health(t *testing.T) {
	t.Run("healthy when initialized", func(t *testing.T) {
		l := New()
		l.Initialize(nil)

		err := l.Health()

		assert.NoError(t, err)
	})

	t.Run("unhealthy when logger is nil", func(t *testing.T) {
		l := &Logger{
			BaseClient: clients.NewBaseClient(ClientName),
			logger:     nil,
		}

		err := l.Health()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

func TestLogger_Shutdown(t *testing.T) {
	l := New()
	l.Initialize(nil)

	err := l.Shutdown()

	assert.NoError(t, err)
	assert.False(t, l.IsInitialized())
}

func TestLogger_GetLogger(t *testing.T) {
	l := New()

	logger := l.GetLogger()

	assert.NotNil(t, logger)
	assert.Equal(t, l.logger, logger)
}

func TestLogger_WithField(t *testing.T) {
	l := New()
	l.Initialize(nil)

	entry := l.WithField("key", "value")

	assert.NotNil(t, entry)
	assert.Equal(t, "value", entry.Data["key"])
}

func TestLogger_WithFields(t *testing.T) {
	l := New()
	l.Initialize(nil)

	fields := map[string]any{
		"field1": "value1",
		"field2": 42,
	}
	entry := l.WithFields(fields)

	assert.NotNil(t, entry)
	assert.Equal(t, "value1", entry.Data["field1"])
	assert.Equal(t, 42, entry.Data["field2"])
}

func TestLogger_WithError(t *testing.T) {
	l := New()
	l.Initialize(nil)

	testErr := errors.New("test error")
	entry := l.WithError(testErr)

	assert.NotNil(t, entry)
	assert.Equal(t, testErr, entry.Data["error"])
}

func TestLogger_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelDebug, Format: FormatText, Output: buf})

	l.Debug("test message")

	assert.Contains(t, buf.String(), "test message")
	assert.Contains(t, strings.ToLower(buf.String()), "debug")
}

func TestLogger_Debugf(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelDebug, Format: FormatText, Output: buf})

	l.Debugf("test %s %d", "message", 42)

	assert.Contains(t, buf.String(), "test message 42")
}

func TestLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelInfo, Format: FormatText, Output: buf})

	l.Info("info message")

	assert.Contains(t, buf.String(), "info message")
	assert.Contains(t, strings.ToLower(buf.String()), "info")
}

func TestLogger_Infof(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelInfo, Format: FormatText, Output: buf})

	l.Infof("info %s", "formatted")

	assert.Contains(t, buf.String(), "info formatted")
}

func TestLogger_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelWarn, Format: FormatText, Output: buf})

	l.Warn("warning message")

	assert.Contains(t, buf.String(), "warning message")
	assert.Contains(t, strings.ToLower(buf.String()), "warn")
}

func TestLogger_Warnf(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelWarn, Format: FormatText, Output: buf})

	l.Warnf("warning %s", "formatted")

	assert.Contains(t, buf.String(), "warning formatted")
}

func TestLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelError, Format: FormatText, Output: buf})

	l.Error("error message")

	assert.Contains(t, buf.String(), "error message")
	assert.Contains(t, strings.ToLower(buf.String()), "error")
}

func TestLogger_Errorf(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelError, Format: FormatText, Output: buf})

	l.Errorf("error %s", "formatted")

	assert.Contains(t, buf.String(), "error formatted")
}

func TestLogger_SetLevel(t *testing.T) {
	l := New()
	l.Initialize(nil)

	l.SetLevel(LevelDebug)
	assert.Equal(t, LevelDebug, l.GetLevel())

	l.SetLevel(LevelError)
	assert.Equal(t, LevelError, l.GetLevel())

	// Invalid level should default to info
	l.SetLevel(Level("invalid"))
	assert.Equal(t, LevelInfo, l.GetLevel())
}

func TestLogger_GetLevel(t *testing.T) {
	l := New()
	l.Initialize(&Config{Level: LevelWarn})

	level := l.GetLevel()

	// logrus returns "warning" not "warn"
	assert.Equal(t, Level("warning"), level)
}

func TestLogger_JSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelInfo, Format: FormatJSON, Output: buf})

	l.WithField("key", "value").Info("json test")

	output := buf.String()
	assert.Contains(t, output, `"key":"value"`)
	assert.Contains(t, output, `"msg":"json test"`)
}

func TestLogger_TextFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelInfo, Format: FormatText, Output: buf})

	l.WithField("key", "value").Info("text test")

	output := buf.String()
	assert.Contains(t, output, "key=value")
	assert.Contains(t, output, "text test")
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, LevelInfo, cfg.Level)
	assert.Equal(t, FormatJSON, cfg.Format)
	assert.NotNil(t, cfg.Output)
}

func TestLogger_LevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New()
	l.Initialize(&Config{Level: LevelWarn, Format: FormatText, Output: buf})

	// These should not appear
	l.Debug("debug message")
	l.Info("info message")

	// These should appear
	l.Warn("warn message")
	l.Error("error message")

	output := buf.String()
	assert.NotContains(t, output, "debug message")
	assert.NotContains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestLogLevelConstants(t *testing.T) {
	assert.Equal(t, Level("debug"), LevelDebug)
	assert.Equal(t, Level("info"), LevelInfo)
	assert.Equal(t, Level("warn"), LevelWarn)
	assert.Equal(t, Level("error"), LevelError)
	assert.Equal(t, Level("fatal"), LevelFatal)
	assert.Equal(t, Level("panic"), LevelPanic)
}

func TestFormatConstants(t *testing.T) {
	assert.Equal(t, Format("json"), FormatJSON)
	assert.Equal(t, Format("text"), FormatText)
}

func TestClientName(t *testing.T) {
	assert.Equal(t, "logger", ClientName)
}
