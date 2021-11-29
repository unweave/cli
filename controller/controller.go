package controller

import (
	"github.com/unweave/cli/api"
	"github.com/unweave/cli/config"
)

type Controller struct {
	api *api.Api
	cfg *config.Config
}

func New() *Controller {
	return &Controller{
		api: api.New(),
		cfg: config.New(),
	}
}
