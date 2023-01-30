package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

var (
	attentionColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#3377FF"))
	successColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#3DB958"))
	errorColor     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E13251"))
)

func Attentionf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(attentionColor.Render(wordwrap.String(s, MaxOutputLineLength)))
}

func Errorf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(errorColor.Render(wordwrap.String(s, MaxOutputLineLength)))
}

func Infof(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(wordwrap.String(s, MaxOutputLineLength))
}

func Successf(format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	fmt.Println(successColor.Render(wordwrap.String(s, MaxOutputLineLength)))
}
