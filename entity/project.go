package entity

import "fmt"

type Project struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (p Project) String() string {
	return fmt.Sprintf("ID: %s, Name: %s", p.Id, p.Name)
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
