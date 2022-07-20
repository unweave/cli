package controller

import (
	"context"

	"github.com/unweave/cli/model"
)

func (c *Controller) GetProjects(ctx context.Context) ([]*model.Project, error) {
	return c.api.GetUserProjects(ctx)
}
