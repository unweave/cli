package controller

import (
	"context"
	"fmt"
	"github.com/unweave/cli/entity"
)

func (c *Controller) Link(ctx context.Context, projectID, path string) error {
	project, err := c.api.GetUserProject(ctx, projectID)
	if err != nil {
		return err
	}

	config := entity.ProjectConfig{
		ID: project.ID,
	}

	err = c.cfg.AddProject(path, config)
	if err != nil {
		return err
	}
	fmt.Printf("Linked project %s with ID %s to path %s\n", project.Name, project.ID, path)
	return nil
}
