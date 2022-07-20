package controller

import (
	"context"

	"github.com/unweave/cli/model"
)

func (c *Controller) GetUserTokens(ctx context.Context) ([]*model.UserToken, error) {
	return c.api.GetUserTokens(ctx)
}

func (c *Controller) CreateUserToken(ctx context.Context) (*model.UserToken, error) {
	return c.api.CreateUserToken(ctx)
}
