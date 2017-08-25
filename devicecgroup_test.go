package devicecgroup

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Device Cgroups", func() {
	var (
		tempDir          string
		cgroupPath       string
		devicesAllowPath string
		devicesDenyPath  string
		devicesListPath  string
	)

	BeforeEach(func() {
		tempDir = createTempDir()
		cgroupPath = filepath.Join(tempDir, "cgroup")
		devicesAllowPath = filepath.Join(cgroupPath, "devices.allow")
		devicesDenyPath = filepath.Join(cgroupPath, "devices.deny")
		devicesListPath = filepath.Join(cgroupPath, "devices.list")
		Expect(os.Mkdir(cgroupPath, os.ModePerm)).To(Succeed())
		createFile(devicesAllowPath)
		createFile(devicesDenyPath)
		createFile(devicesListPath)
	})

	AfterEach(func() {
		if strings.HasPrefix(tempDir, os.TempDir()) {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}
	})

	Describe("#Load", func() {
		It("does not return an error", func() {
			_, err := Load(cgroupPath)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when path does not exist", func() {
			It("returns an error", func() {
				_, err := Load("")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when path is not a device cgroup", func() {
			It("returns an error", func() {
				_, err := Load(tempDir)
				expectedErr := fmt.Errorf("not a device cgroup: %s", tempDir)
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})

	Describe("#List", func() {
		var deviceCgroup DeviceCgroup

		BeforeEach(func() {
			var err error
			deviceCgroup, err = Load(cgroupPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(ioutil.WriteFile(devicesListPath, []byte("c 1:1 rwm"), os.ModePerm))
		})

		It("does not return an error", func() {
			_, err := deviceCgroup.List()
			Expect(err).NotTo(HaveOccurred())
		})

		It("lists device properties", func() {
			entries, _ := deviceCgroup.List()
			Expect(entries).To(Equal([]string{"c 1:1 rwm"}))
		})

		Context("when there are multiple device entries", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(devicesListPath, []byte("c 1:1 rwm\nb *:* m"), os.ModePerm))
			})

			It("lists device properties", func() {
				entries, _ := deviceCgroup.List()
				Expect(entries).To(Equal([]string{"c 1:1 rwm", "b *:* m"}))
			})
		})

		Context("when reading the device list returns an error", func() {
			BeforeEach(func() {
				Expect(strings.HasPrefix(devicesListPath, os.TempDir())).To(BeTrue())
				Expect(os.Remove(devicesListPath)).To(Succeed())
			})

			It("returns the error", func() {
				_, err := deviceCgroup.List()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when there are no device entries", func() {
			BeforeEach(func() {
				Expect(os.Truncate(devicesListPath, 0)).To(Succeed())
			})

			It("does not return an error", func() {
				_, err := deviceCgroup.List()
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an empty device list", func() {
				entries, _ := deviceCgroup.List()
				Expect(entries).To(BeEmpty())
			})
		})
	})

	Describe("#Allow", func() {
		var deviceCgroup DeviceCgroup

		BeforeEach(func() {
			var err error
			deviceCgroup, err = Load(cgroupPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return an error", func() {
			Expect(deviceCgroup.Allow("c *:* m")).To(Succeed())
		})

		It("writes the entry to devices.allow", func() {
			deviceCgroup.Allow("c *:* m")
			Expect(readFile(devicesAllowPath)).To(Equal("c *:* m\n"))
		})

		Context("when allowing no devices", func() {
			It("does not return an error", func() {
				Expect(deviceCgroup.Allow()).To(Succeed())
			})

			It("writes nothing to devices.allow", func() {
				deviceCgroup.Allow()
				Expect(readFile(devicesAllowPath)).To(BeEmpty())
			})
		})

		Context("when allowing three devices", func() {
			It("does not return an error", func() {
				Expect(deviceCgroup.Allow("c *:* m", "b *:* m", "c 10:229 rwm")).To(Succeed())
			})

			It("writes the entries to devices.allow", func() {
				deviceCgroup.Allow("c *:* m", "b *:* m", "c 10:229 rwm")
				Expect(readFile(devicesAllowPath)).To(Equal("c *:* m\nb *:* m\nc 10:229 rwm\n"))
			})
		})

		Context("when opening devices.allow returns an error", func() {
			BeforeEach(func() {
				Expect(strings.HasPrefix(devicesAllowPath, os.TempDir())).To(BeTrue())
				Expect(os.Remove(devicesAllowPath)).To(Succeed())
			})

			It("returns the error", func() {
				Expect(deviceCgroup.Allow()).NotTo(Succeed())
			})
		})

		Context("when writing to devices.allow returns an error", func() {
			BeforeEach(func() {
				deviceCgroup.fprinter = func(w io.Writer, a ...interface{}) (int, error) {
					return 0, errors.New("oh my god")
				}
			})

			It("returns the error", func() {
				err := deviceCgroup.Allow("c *:* m")
				Expect(err).To(MatchError("oh my god"))
			})
		})
	})

	Describe("#Deny", func() {
		var deviceCgroup DeviceCgroup

		BeforeEach(func() {
			var err error
			deviceCgroup, err = Load(cgroupPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return an error", func() {
			Expect(deviceCgroup.Deny("c *:* m")).To(Succeed())
		})

		It("writes the entry to devices.deny", func() {
			deviceCgroup.Deny("c *:* m")
			Expect(readFile(devicesDenyPath)).To(Equal("c *:* m\n"))
		})

		Context("when denying no devices", func() {
			It("does not return an error", func() {
				Expect(deviceCgroup.Deny()).To(Succeed())
			})

			It("writes nothing to devices.deny", func() {
				deviceCgroup.Deny()
				Expect(readFile(devicesDenyPath)).To(BeEmpty())
			})
		})

		Context("when denying three devices", func() {
			It("does not return an error", func() {
				Expect(deviceCgroup.Deny("c *:* m", "b *:* m", "c 10:229 rwm")).To(Succeed())
			})

			It("writes the entries to devices.deny", func() {
				deviceCgroup.Deny("c *:* m", "b *:* m", "c 10:229 rwm")
				Expect(readFile(devicesDenyPath)).To(Equal("c *:* m\nb *:* m\nc 10:229 rwm\n"))
			})
		})

		Context("when opening devices.deny returns an error", func() {
			BeforeEach(func() {
				Expect(strings.HasPrefix(devicesDenyPath, os.TempDir())).To(BeTrue())
				Expect(os.Remove(devicesDenyPath)).To(Succeed())
			})

			It("returns the error", func() {
				Expect(deviceCgroup.Deny()).NotTo(Succeed())
			})
		})

		Context("when writing to devices.deny returns an error", func() {
			BeforeEach(func() {
				deviceCgroup.fprinter = func(w io.Writer, a ...interface{}) (int, error) {
					return 0, errors.New("oh my god")
				}
			})

			It("returns the error", func() {
				err := deviceCgroup.Deny("c *:* m")
				Expect(err).To(MatchError("oh my god"))
			})
		})
	})

	Describe("#HasChildren", func() {
		var deviceCgroup DeviceCgroup

		BeforeEach(func() {
			var err error
			deviceCgroup, err = Load(cgroupPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return an error", func() {
			_, err := deviceCgroup.HasChildren()
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns false", func() {
			has, _ := deviceCgroup.HasChildren()
			Expect(has).To(BeFalse())
		})

		Context("when child cgroups exist", func() {
			var childCgroupPath string

			BeforeEach(func() {
				childCgroupPath = filepath.Join(cgroupPath, "nested")
				nestedDevicesAllowPath := filepath.Join(childCgroupPath, "devices.allow")
				nesteddevicesDenyPath := filepath.Join(childCgroupPath, "devices.deny")
				nesteddevicesListPath := filepath.Join(childCgroupPath, "devices.list")
				Expect(os.Mkdir(childCgroupPath, os.ModePerm)).To(Succeed())
				createFile(nestedDevicesAllowPath)
				createFile(nesteddevicesDenyPath)
				createFile(nesteddevicesListPath)
			})

			It("does not return an error", func() {
				_, err := deviceCgroup.HasChildren()
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true", func() {
				has, _ := deviceCgroup.HasChildren()
				Expect(has).To(BeTrue())
			})

			Context("when ReadDir returns an error", func() {
				BeforeEach(func() {
					deviceCgroup.dirReader = func(_ string) ([]os.FileInfo, error) {
						return nil, errors.New(":(")
					}
				})

				It("returns the error", func() {
					_, err := deviceCgroup.HasChildren()
					Expect(err).To(MatchError(":("))
				})
			})
		})
	})
})

func createTempDir() string {
	dir, err := ioutil.TempDir("", "")
	Expect(err).NotTo(HaveOccurred())
	return dir
}

func createFile(path string) {
	file, err := os.Create(path)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())
}

func readFile(path string) string {
	content, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return string(content)
}
