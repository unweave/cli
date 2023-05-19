package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/franela/goblin"
)

func TestParseSSHArgs(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("parseSSHArgs", func() {
		g.It("should return an empty map for no arguments", func() {
			args := []string{}
			result, err := parseSSHArgs(args)
			g.Assert(err).Equal(nil)
			g.Assert(len(result)).Equal(0)
		})

		g.It("should parse execRef when there is only one argument", func() {
			execRef := "example-execRef"
			args := []string{execRef}
			result, err := parseSSHArgs(args)
			g.Assert(err).Equal(nil)
			g.Assert(len(result)).Equal(1)
			g.Assert(result["execRef"]).Equal(execRef)
		})

		g.It("should dynamically parse SSH arguments with flags and values", func() {
			args := []string{"--username", "john", "--port", "22", "--forward-agent"}
			result, err := parseSSHArgs(args)
			g.Assert(err).Equal(nil)
			g.Assert(len(result)).Equal(3)
			g.Assert(result["username"]).Equal("john")
			g.Assert(result["port"]).Equal("22")
			g.Assert(result["forward-agent"]).Equal("")
		})

		g.It("should return an error for invalid argument format", func() {
			args := []string{"--username", "john", "port", "22"}
			result, err := parseSSHArgs(args)
			g.Assert(result).Equal(nil)
			g.Assert(err).Equal(errors.New("invalid arguments format. Use '--' as the separator"))
			g.Assert(strings.Contains(err.Error(), "separator")).Equal(true)
		})

		g.It("should return an error for missing value for a flag", func() {
			args := []string{"--username", "john", "--port"}
			result, err := parseSSHArgs(args)
			g.Assert(result).Equal(nil)
			g.Assert(err).Equal(errors.New("invalid arguments format. Missing value for flag: port"))
			g.Assert(strings.Contains(err.Error(), "port")).Equal(true)
		})
	})
}
