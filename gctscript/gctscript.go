package gctscript

import (
	"github.com/openware/gocryptotrader/gctscript/modules"
	"github.com/openware/gocryptotrader/gctscript/wrappers/gct"
)

// Setup configures the wrapper interface to use
func Setup() {
	modules.SetModuleWrapper(gct.Setup())
}
