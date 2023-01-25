package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/unweave/unweave/api/types"
)

type Error struct {
	*types.HTTPError
}

func (e *Error) Short() string {
	str := fmt.Sprintf("%s API error: %s", e.Provider, e.Message)
	return str
}

func (e *Error) Verbose() string {
	header := "API error:\n"
	if e.Provider != "" {
		header = fmt.Sprintf("%s API error:\n", e.Provider.DisplayName())
	}
	body := ""
	if e.Code != 0 {
		s := fmt.Sprintf("Code:       %d", e.Code)
		body += wordwrap.String(s, MaxOutputLineLength-IndentWidth)
		body += "\n"
	}
	if e.Message != "" {
		s := fmt.Sprintf("Message:    %s", e.Message)
		body += wordwrap.String(s, MaxOutputLineLength-IndentWidth)
		body += "\n"
	}
	if e.Suggestion != "" {
		s := fmt.Sprintf("Suggestion: %s", e.Suggestion)
		body += wordwrap.String(s, MaxOutputLineLength-IndentWidth)
		body += "\n"
	}
	str := header + indent.String(body, IndentWidth)
	return str
}

func HandleError(err error) error {
	var e *types.HTTPError
	if errors.As(err, &e) {
		if e.Code == 401 {
			fmt.Println("Unauthorized. Please login with `unweave login`")
			os.Exit(1)
			return nil
		}
		uie := &Error{HTTPError: e}
		fmt.Println(uie.Verbose())
		os.Exit(1)
	}
	return err
}
