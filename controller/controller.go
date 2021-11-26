package controller

import (
	"github.com/unweave/cli/api"
)

type Controller struct {
	api *api.Api
}

func New() *Controller {
	return &Controller{
		api: api.New(),
	}
}
