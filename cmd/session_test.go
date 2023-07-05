package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unweave/cli/config"
	"github.com/unweave/unweave/api/types"
)

func TestParseHardwareSpec(t *testing.T) {
	t.Run("should use default config and merge in custom fields", func(t *testing.T) {
		f, err := os.CreateTemp("", "unweave-config-*.toml")
		assert.NoError(t, err)

		f.WriteString(`
[[specs]]
name = "default"
[specs.cpu]
type = "x86_64"
`)
		f.Sync()

		secrets, project := config.InitProjectConfigFrom(f.Name(), "")
		config.Config.Project = project
		config.Config.Project.Env = secrets

		config.CPUs = 1
		config.GPUType = "rtx_4000"
		config.GPUs = 2
		config.GPUMemory = 3
		config.Memory = 4
		config.HDD = 5

		hws, err := parseHardwareSpec()
		assert.NoError(t, err)

		want := types.HardwareSpec{
			CPU: types.CPU{Type: "x86_64", HardwareRequestRange: hwRange(1, 1)},
			GPU: types.GPU{Type: "rtx_4000", Count: hwRange(2, 2), RAM: hwRange(3, 3)},
			RAM: hwRange(4, 4),
			HDD: hwRange(5, 5),
		}
		assert.Equal(t, want, hws)
	})

	t.Run("should use custom spec and merge in fields", func(t *testing.T) {
		f, err := os.CreateTemp("", "unweave-config-*.toml")
		assert.NoError(t, err)

		f.WriteString(`
[[specs]]
name = "my-spec"
[specs.cpu]
type = "x86_64"
[specs.gpu]
type = "abc_123"
count = 4
memory = 8
[specs.hdd]
size = 10
`)
		f.Sync()

		secrets, project := config.InitProjectConfigFrom(f.Name(), "")
		config.Config.Project = project
		config.Config.Project.Env = secrets

		// reset to zero
		config.CPUs = 0
		config.GPUType = ""
		config.GPUs = 0
		config.GPUMemory = 0
		config.Memory = 0
		config.HDD = 0

		// set some values
		config.CPUs = 1
		config.Memory = 4
		config.SpecName = "my-spec"

		hws, err := parseHardwareSpec()
		assert.NoError(t, err)

		want := types.HardwareSpec{
			CPU: types.CPU{Type: "x86_64", HardwareRequestRange: hwRange(1, 1)},
			GPU: types.GPU{Type: "abc_123", Count: hwRange(4, 4), RAM: hwRange(8, 8)},
			RAM: hwRange(4, 4),
			HDD: hwRange(10, 10),
		}
		assert.Equal(t, want, hws)
	})
}

func hwRange(min, max int) types.HardwareRequestRange {
	return types.HardwareRequestRange{Min: min, Max: max}
}
