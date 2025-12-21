package metadata

// Metadata represents application metadata for CLI and runtime use
type Metadata interface {
	Name() string    // App name (e.g., "caop")
	Short() string   // Short description (e.g., "CAOP CLI")
	Long() string    // Long description (e.g., "CAOP - Commerce and Online Purchases")
	Version() string // App version (e.g., "v1.2.3" or "dev")
}

// Info is a simple implementation of Metadata
type Info struct {
	AppName    string
	AppShort   string
	AppLong    string
	AppVersion string
}

func (i Info) Name() string    { return i.AppName }
func (i Info) Short() string   { return i.AppShort }
func (i Info) Long() string    { return i.AppLong }
func (i Info) Version() string { return i.AppVersion }

// Default returns the framework's default metadata
func Default() Metadata {
	return Info{
		AppName:    "codo",
		AppShort:   "Codo Framework CLI",
		AppLong:    "Codo Framework - A production-ready Go backend framework",
		AppVersion: "dev",
	}
}
