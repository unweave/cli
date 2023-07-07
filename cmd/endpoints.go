package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
)

func EndpointCreate(cmd *cobra.Command, args []string) error {
	execID := args[0]
	ctx := cmd.Context()

	owner, projectName := config.GetProjectOwnerAndName()

	uwc := config.InitUnweaveClient()

	endpoint, err := uwc.Endpoints.Create(ctx, owner, projectName, execID)
	if err != nil {
		return err
	}

	ui.Infof("endpoint: %s", endpoint.ID)
	return nil
}

func EndpointList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	owner, projectName := config.GetProjectOwnerAndName()

	uwc := config.InitUnweaveClient()

	endpoints, err := uwc.Endpoints.List(ctx, owner, projectName)
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		ui.Infof("id: %s", endpoint.ID)
		ui.Infof("exec id: %s", endpoint.ExecID)
		ui.Infof("http: %s\n", endpoint.HTTPEndpoint)
	}

	return nil
}

func EndpointEvalAttach(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	endpointID := args[0]
	evalID := args[1]

	owner, projectName := config.GetProjectOwnerAndName()

	uwc := config.InitUnweaveClient()

	err := uwc.Endpoints.EvalAttach(ctx, owner, projectName, endpointID, evalID)
	if err != nil {
		return err
	}

	return nil
}

func EndpointEvalCheck(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	endpointID := args[0]

	owner, projectName := config.GetProjectOwnerAndName()

	uwc := config.InitUnweaveClient()

	err := uwc.Endpoints.RunEvalCheck(ctx, owner, projectName, endpointID)
	if err != nil {
		return err
	}

	return nil
}
