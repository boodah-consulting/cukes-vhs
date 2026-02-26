package cukesvhs

import (
	"github.com/spf13/afero"
)

// DefaultFs returns the default OS filesystem for production use.
// Tests can use afero.NewMemMapFs() for isolation.
func DefaultFs() afero.Fs {
	return afero.NewOsFs()
}
