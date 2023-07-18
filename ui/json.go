package ui

import (
	"encoding/json"
	"os"
)

func JSON(v any) {
	if !OutputJSON {
		return
	}

	err := json.NewEncoder(os.Stdout).Encode(v)
	if err != nil {
		Errorf("failed to encode output: %s", err.Error())
	}
}
