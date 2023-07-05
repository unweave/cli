package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecs(t *testing.T) {
	t.Parallel()

	t.Run("should load specs", func(t *testing.T) {
		t.Parallel()
		fixture := `
default_provider = "unweave"
[[specs]]
name = "default"
[specs.cpu]
type = "x86_64"
count = 3
memory = 4
[specs.gpu]
type = "rtx_4000"
count = 2
memory = 8
[specs.hdd]
size = 10
[[specs]]
name = "my-spec"
[specs.cpu]
type = "x86_64"
[specs.gpu]
type = "rtx_5000"
count = 1
memory = 4
`
		f, err := os.CreateTemp("", "unweave-config-*.toml")
		if err != nil {
			t.Error(err)
		}

		_, err = f.WriteString(fixture)
		if err != nil {
			t.Error(err)
		}
		f.Sync()

		var p project
		if err := readAndUnmarshal(f.Name(), &p); err != nil {
			t.Error(err)
		}

		want :=
			project{
				DefaultProvider: "unweave",
				Specs: []spec{
					{
						CPU: specResources{Type: "x86_64", Count: 3, Memory: 4},
						GPU: specResources{Type: "rtx_4000", Count: 2, Memory: 8},
						HDD: specHDD{Size: 10},
					},
					{
						CPU: specResources{Type: "x86_64"},
						GPU: specResources{Type: "rtx_5000", Count: 1, Memory: 4},
					},
				},
			}

		assert.Equal(t, want, p)

	})
}
