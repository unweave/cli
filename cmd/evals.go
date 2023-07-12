package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
)

func EvalCreate(cmd *cobra.Command, args []string) error {
	execID := args[0]
	ctx := cmd.Context()

	owner, projectName := config.GetProjectOwnerAndName()

	uwc := config.InitUnweaveClient()

	eval, err := uwc.Evals.Create(ctx, owner, projectName, execID)
	if err != nil {
		return err
	}

	ui.Infof("eval: %s", eval.ID)
	return nil
}

func EvalList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	owner, projectName := config.GetProjectOwnerAndName()

	uwc := config.InitUnweaveClient()

	evals, err := uwc.Evals.List(ctx, owner, projectName)
	if err != nil {
		return err
	}

	for _, eval := range evals {
		ui.Infof("id: %s", eval.ID)
		ui.Infof("exec id: %s", eval.ExecID)
		ui.Infof("http endpoint: %s\n", eval.HTTPEndpoint)
	}

	return nil
}
