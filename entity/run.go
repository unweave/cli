package entity

import (
	"io"
)

type Zepl struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Command       string `json:"command"`
	ProcessorType string `json:"processorType"`
	CurrentStatus string `json:"currentStatus"`
}

type GatherContextFunc func(w io.Writer) error

const InitZeplMutation = `
	mutation InitZepl ($projectID: UUID!, $command: String!, $gpu: Boolean) {
		initZepl (projectID: $projectID, command: $command, gpu: $gpu) {
			id
			name
			command
			processorType
			currentStatus
		}
	}
`
