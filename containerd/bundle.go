package containerd

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// incompatibleLinuxMounts is the set of default mounts generated by containerd
// for Linux.  These mounts are not valid on FreeBSD.
var incompatibleLinuxMounts = map[string]specs.Mount{
	"/proc": {
		Destination: "/proc",
		Type:        "proc",
		Source:      "proc",
		Options:     []string{"nosuid", "noexec", "nodev"},
	},
	"/dev": {
		Destination: "/dev",
		Type:        "tmpfs",
		Source:      "tmpfs",
		Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
	},
	"/dev/pts": {
		Destination: "/dev/pts",
		Type:        "devpts",
		Source:      "devpts",
		Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
	},
	"/dev/shm": {
		Destination: "/dev/shm",
		Type:        "tmpfs",
		Source:      "shm",
		Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
	},
	"/dev/mqueue": {
		Destination: "/dev/mqueue",
		Type:        "mqueue",
		Source:      "mqueue",
		Options:     []string{"nosuid", "noexec", "nodev"},
	},
	"/sys": {
		Destination: "/sys",
		Type:        "sysfs",
		Source:      "sysfs",
		Options:     []string{"nosuid", "noexec", "nodev", "ro"},
	},
	"/run": {
		Destination: "/run",
		Type:        "tmpfs",
		Source:      "tmpfs",
		Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
	},
}

// filterIncompatibleLinuxMounts removes Linux-specific default mounts that
// might be present in a containerd-generated OCI bundle
func filterIncompatibleLinuxMounts(bundle string) error {
	if bundle == "" {
		return nil
	}
	configJSON := filepath.Join(bundle, "config.json")
	spec := &specs.Spec{}
	configBytes, err := ioutil.ReadFile(configJSON)
	if err != nil {
		return err
	}
	err = json.Unmarshal(configBytes, spec)
	if err != nil {
		return err
	}

	mounts := make([]specs.Mount, 0)
	for _, m := range spec.Mounts {
		if toFilter, ok := incompatibleLinuxMounts[m.Destination]; ok {
			if equalMounts(m, toFilter) {
				continue
			}
		}
		mounts = append(mounts, m)
	}
	if len(spec.Mounts) == len(mounts) {
		return nil
	}
	spec.Mounts = mounts
	out, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configJSON, out, 0)
}

// equalMounts compares two mounts to determine whether they are equal
func equalMounts(a, b specs.Mount) bool {
	if a.Source != b.Source ||
		a.Destination != b.Destination ||
		a.Type != b.Type ||
		len(a.Options) != len(b.Options) {
		return false
	}
	for i := 0; i < len(a.Options); i++ {
		if a.Options[i] != b.Options[i] {
			return false
		}
	}
	return true
}
