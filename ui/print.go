package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/unweave/cli/vars"
	"github.com/unweave/unweave/api/types"
)

var (
	attentionColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5C237"))
	successColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#3DB958"))
	errorColor     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E13251"))
	warningColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5C237"))
)

func Attentionf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(attentionColor.Render(wordwrap.String(s, MaxOutputLineLength)))
}

func Errorf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(errorColor.Render(wordwrap.String(s, MaxOutputLineLength)))
}

// Fatal prints an error message. It checks is the error is a types.Error and prints
// the verbose message if it is.
//
// You should use this function instead of Errorf (which is being deprecated).
func Fatal(msg string, err error) {
	fmt.Println(errorColor.Render(wordwrap.String(msg, MaxOutputLineLength)))
	var e *types.Error
	if errors.As(err, &e) {
		if e.Code == 401 {
			fmt.Println("Unauthorized. Please login with `unweave login`")
			os.Exit(1)
		}
		uie := &Error{Error: e}
		fmt.Println(uie.Verbose())
	}
	os.Exit(1)
}

func Infof(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(wordwrap.String(s, MaxOutputLineLength))
}

func Successf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(successColor.Render(wordwrap.String(s, MaxOutputLineLength)))
}

func Debugf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	if vars.Debug {
		fmt.Println(wordwrap.String(s, MaxOutputLineLength))
	}
}
