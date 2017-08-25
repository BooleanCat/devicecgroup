package devicecgroup_system_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/BooleanCat/devicecgroup"
)

const testDeviceCGroupPath = "/sys/fs/cgroup/devices/test"

var _ = Describe("DeviceCgroup", func() {
	var (
		deviceCgroup devicecgroup.DeviceCgroup
		loadErr      error
	)

	BeforeEach(func() {
		Expect(os.Mkdir(testDeviceCGroupPath, os.ModePerm)).To(Succeed())
		deviceCgroup, loadErr = devicecgroup.Load(testDeviceCGroupPath)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(testDeviceCGroupPath)).To(Succeed())
	})

	It("loads without error", func() {
		Expect(loadErr).NotTo(HaveOccurred())
	})

	Context("when allowing all devices", func() {
		BeforeEach(func() {
			Expect(deviceCgroup.Allow("a")).To(Succeed())
		})

		It("can use list to confirm devices are allowed", func() {
			entries, err := deviceCgroup.List()
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(Equal([]string{"a *:* rwm"}))
		})
	})

	Context("when denying all devices", func() {
		BeforeEach(func() {
			Expect(deviceCgroup.Deny("a")).To(Succeed())
		})

		It("can use list to confirm devices are allowed", func() {
			entries, err := deviceCgroup.List()
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(BeEmpty())
		})
	})

	Context("when allowing only /dev/zero", func() {
		var allowErr error

		BeforeEach(func() {
			Expect(deviceCgroup.Deny("a")).To(Succeed())
			allowErr = deviceCgroup.Allow("c 1:5 rwm")
		})

		It("does not return an error", func() {
			Expect(allowErr).NotTo(HaveOccurred())
		})

		It("lists only /dev/zero", func() {
			entries, err := deviceCgroup.List()
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(Equal([]string{"c 1:5 rwm"}))
		})
	})

	Context("when allowing only /dev/zero, /dev/null and /dev/random", func() {
		BeforeEach(func() {
			Expect(deviceCgroup.Deny("a")).To(Succeed())
			Expect(deviceCgroup.Allow(
				"c 1:5 rwm",
				"c 1:3 rwm",
				"c 1:8 rwm",
			)).To(Succeed())
		})

		It("lists /dev/zero, /dev/null and /dev/random", func() {
			entries, err := deviceCgroup.List()
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(Equal([]string{"c 1:5 rwm", "c 1:3 rwm", "c 1:8 rwm"}))
		})

		Context("and then denying /dev/zero and /dev/random", func() {
			BeforeEach(func() {
				Expect(deviceCgroup.Deny("c 1:5 rwm", "c 1:8 rwm")).To(Succeed())
			})

			It("lists /dev/null", func() {
				entries, err := deviceCgroup.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(entries).To(Equal([]string{"c 1:3 rwm"}))
			})
		})
	})

	Context("when a devicecgroup has children", func() {
		var nestedDeviceCgroupPath = "/sys/fs/cgroup/devices/test/nested"

		BeforeEach(func() {
			Expect(os.Mkdir(nestedDeviceCgroupPath, os.ModePerm)).To(Succeed())
		})

		It("loads without error", func() {
			Expect(loadErr).NotTo(HaveOccurred())
		})

		It("can tell if there are child cgroups", func() {
			hasChildren, err := deviceCgroup.HasChildren()
			Expect(err).NotTo(HaveOccurred())
			Expect(hasChildren).To(BeTrue())
		})
	})
})

func devicecgroupLoad(path string) devicecgroup.DeviceCgroup {
	deviceCgroup, err := devicecgroup.Load(path)
	Expect(err).NotTo(HaveOccurred())
	return deviceCgroup
}
