package model

import "fmt"

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p Project) String() string {
	return fmt.Sprintf("ID: %s, Name: %s", p.ID, p.Name)
}

const GetProjectQuery = `
	query GetProject ($id: UUID!) {
		project (id: $id) {
			id
			name
		}
	}
`

const GetProjectsQuery = `
	query GetProjects {
		projects {
			id
			name
		}
	}
`
