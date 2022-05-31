package controller

import (
	"context"
	"github.com/unweave/cli/entity"
)

func (c *Controller) GetUserTokens(ctx context.Context) ([]entity.UserToken, error) {
	return c.api.GetUserTokens(ctx)
}

func (c *Controller) CreateUserToken(ctx context.Context) (*entity.UserToken, error) {
	return c.api.CreateUserToken(ctx)
}
