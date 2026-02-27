package cukesvhs

import (
	"github.com/spf13/afero"
)

var defaultFs afero.Fs = afero.NewOsFs()

// SetDefaultFs allows the CLI or tests to override the default filesystem.
func SetDefaultFs(fs afero.Fs) {
	defaultFs = fs
}

// DefaultFs returns the default filesystem for production use.
// Tests and CLI can override this using SetDefaultFs.
func DefaultFs() afero.Fs {
	return defaultFs
}
