// Package auto provides automatic registration of all framework CLI commands.
// Import this package with a blank identifier to register all commands:
//
//	import _ "github.com/codoworks/codo-framework/cmd/auto"
package auto

import (
	// Register all framework commands via their init() functions
	_ "github.com/codoworks/codo-framework/cmd/db"
	_ "github.com/codoworks/codo-framework/cmd/info"
	_ "github.com/codoworks/codo-framework/cmd/start"
	_ "github.com/codoworks/codo-framework/cmd/task"
)
