package entity

import (
	"io"
)

type Zepl struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Command       string `json:"command"`
	CurrentStatus string `json:"currentStatus"`
}

type GatherContextFunc func(w io.Writer) error

const InitZeplMutation = `
	mutation InitZepl ($projectID: UUID!, $command: String!) {
		initZepl (projectID: $projectID, command: $command) {
			id
			name
			command
			currentStatus
		}
	}
`
