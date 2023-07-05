package config

import (
	"testing"

	. "github.com/franela/goblin"
)

func TestProjectHostDir(t *testing.T) {
	g := Goblin(t)

	g.Describe("ProjectHostDir", func() {
		g.It("should append the root directory to UnweaveHostDir", func() {
			g.Assert(ProjectHostDir()).Equal("/home/unweave")
		})
	})
}
