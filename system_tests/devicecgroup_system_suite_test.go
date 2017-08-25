package devicecgroup_system_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDevicecgroupSystem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Devicecgroup System Suite")
}
