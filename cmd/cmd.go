package cmd

import (
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/controller"
)

type Handler struct {
	ctrl *controller.Controller
	cfg *config.Config
}


func New() *Handler {
	return &Handler{
		ctrl: controller.New(),
		cfg:  config.New(),
	}
}