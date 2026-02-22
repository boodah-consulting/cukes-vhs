package vhsgen_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVhsgen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vhsgen Suite")
}
