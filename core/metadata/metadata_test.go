package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	meta := Default()
	assert.Equal(t, "codo", meta.Name())
	assert.Equal(t, "Codo Framework CLI", meta.Short())
	assert.Equal(t, "Codo Framework - A production-ready Go backend framework", meta.Long())
}

func TestInfo(t *testing.T) {
	meta := Info{
		AppName:  "testapp",
		AppShort: "Test CLI",
		AppLong:  "Test Application",
	}
	assert.Equal(t, "testapp", meta.Name())
	assert.Equal(t, "Test CLI", meta.Short())
	assert.Equal(t, "Test Application", meta.Long())
}
