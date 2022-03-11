package entity

import (
	"io"
)

type Zepl struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Command       string `json:"command"`
	InstanceType  string `json:"instance_type"`
	CurrentStatus string `json:"currentStatus"`
}

type GatherContextFunc func(w io.Writer) error

const InitZeplMutation = `
	mutation InitZepl ($projectID: UUID!, $command: String!, $instanceType: String) {
		initZepl (projectID: $projectID, command: $command, instanceType: $instanceType) {
			id
			name
			command
			instanceType
			currentStatus
		}
	}
`
