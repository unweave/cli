package ui

import (
	"os"

	"github.com/manifoldco/promptui"
)

type bellSkipper struct{}

// Write implements an io.WriterCloser over os.Stderr, but it skips the terminal
// bell character.
func (bs *bellSkipper) Write(b []byte) (int, error) {
	const charBell = 7 // c.f. readline.CharBell
	if len(b) == 1 && b[0] == charBell {
		return 0, nil
	}
	return os.Stderr.Write(b)
}

// Close implements an io.WriterCloser over os.Stderr.
func (bs *bellSkipper) Close() error {
	return os.Stderr.Close()
}

func Select(label string, options []string) (int, error) {
	prompt := promptui.Select{
		Label:  label,
		Items:  options,
		Stdout: &bellSkipper{},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return -1, err
	}
	return idx, nil
}
