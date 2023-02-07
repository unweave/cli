package ui

import (
	"github.com/manifoldco/promptui"
)

func Select(label string, options []string) (int, error) {
	prompt := promptui.Select{
		Label: label,
		Items: options,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return -1, err
	}
	return idx, nil
}
