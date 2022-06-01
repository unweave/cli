package controller

import (
	"context"
	"github.com/unweave/cli/entity"
)

func (c *Controller) GetProjects(ctx context.Context) ([]*entity.Project, error) {
	return c.api.GetUserProjects(ctx)
}
