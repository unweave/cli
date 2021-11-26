package controller

import (
	"github.com/unweave/cli/api"
)

type Controller struct {
	gtwy *api.Api
}

func New() *Controller {
	return &Controller{
		gtwy: api.New(),
	}
}
