package devicecgroup_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDevicecgroup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devicecgroup Suite")
}
