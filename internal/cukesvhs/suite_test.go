package cukesvhs_test

import (
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVhsgen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vhsgen Suite")
}

// skipIfWindows skips a test when running on Windows.
// Use for tests that rely on Unix-specific behaviour (permissions, shell scripts).
func skipIfWindows(reason string) {
	if runtime.GOOS == "windows" {
		Skip(reason)
	}
}
