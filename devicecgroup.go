package devicecgroup

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//DeviceCgroup provides golang hooks for Linux device cgroups
type DeviceCgroup struct {
	path      string
	fprinter  func(io.Writer, ...interface{}) (int, error)
	dirReader func(dirname string) ([]os.FileInfo, error)
}

//Load an existing device cgroup at path
func Load(path string) (DeviceCgroup, error) {
	if err := validateDeviceCgroup(path); err != nil {
		return DeviceCgroup{}, err
	}

	return DeviceCgroup{path: path, fprinter: fmt.Fprint, dirReader: ioutil.ReadDir}, nil
}

//List device properties from `devices.list`
func (d DeviceCgroup) List() ([]string, error) {
	content, err := ioutil.ReadFile(d.getDevicesListPath())
	if err != nil {
		return nil, err
	} else if len(content) == 0 {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(string(content)), "\n"), nil
}

//Allow writes lines to `devices.allow`
func (d DeviceCgroup) Allow(entries ...string) error {
	file, err := os.OpenFile(d.getDevicesAllowPath(), os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return d.writeEntryTo(file, entries...)
}

//Deny writes lines to `devices.deny`
func (d DeviceCgroup) Deny(entries ...string) error {
	file, err := os.OpenFile(d.getDevicesDenyPath(), os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return d.writeEntryTo(file, entries...)
}

//HasChildren returns true when a device cgroup has children
func (d DeviceCgroup) HasChildren() (bool, error) {
	dirs, err := d.dirReader(d.path)
	if err != nil {
		return false, err
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

func (d DeviceCgroup) writeEntryTo(writer io.Writer, entries ...string) error {
	for _, entry := range entries {
		_, err := d.fprinter(writer, entry+"\n")
		if err != nil {
			return err
		}
	}

	return nil
}

func (d DeviceCgroup) getDevicesAllowPath() string {
	return filepath.Join(d.path, "devices.allow")
}

func (d DeviceCgroup) getDevicesDenyPath() string {
	return filepath.Join(d.path, "devices.deny")
}

func (d DeviceCgroup) getDevicesListPath() string {
	return filepath.Join(d.path, "devices.list")
}

func validateDeviceCgroup(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	if !containsFile(files, "devices.allow") ||
		!containsFile(files, "devices.deny") ||
		!containsFile(files, "devices.list") {
		return fmt.Errorf("not a device cgroup: %s", path)
	}

	return nil
}

func containsFile(files []os.FileInfo, name string) bool {
	for _, fileInfo := range files {
		if fileInfo.Name() == name {
			return true
		}
	}

	return false
}
