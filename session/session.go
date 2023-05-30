package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

// Create attempts to create a session using the Exec spec provided, uses GPUs in the config if not, returns a 503 out-of-capacity error.
// Renders newly created sessions to the UI implicitly.
func Create(ctx context.Context, params types.ExecCreateParams) (string, error) {
	if params.HardwareSpec.GPU.Type == "" {
		exec, err := createSessionFromConfigGPUTypes(ctx, params)
		renderSessionCreated(exec)

		return exec.ID, err
	}

	exec, err := createSession(ctx, params, params.HardwareSpec.GPU.Type)
	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	renderSessionCreated(exec)

	return exec.ID, err
}

func createSession(ctx context.Context, params types.ExecCreateParams, gpuType string) (*types.Exec, error) {
	uwc := config.InitUnweaveClient()
	owner, projectName := config.GetProjectOwnerAndName()

	useParams := params
	useParams.HardwareSpec.GPU.Type = gpuType

	exec, err := uwc.Exec.Create(ctx, owner, projectName, useParams)
	if err != nil {
		return nil, err
	}

	return exec, nil
}

func createSessionFromConfigGPUTypes(ctx context.Context, params types.ExecCreateParams) (*types.Exec, error) {
	gpuTypesFromConfig := gpuTypesFromConfig()
	var err error
	var exec *types.Exec
	for _, gpuType := range gpuTypesFromConfig {
		exec, err = createSession(ctx, params, gpuType)
		if err != nil {
			if isOutOfCapacityError(err) {
				continue
			}
			return nil, err
		}

		return exec, nil
	}

	return nil, err
}

func isOutOfCapacityError(err error) bool {
	var e *types.Error
	if errors.As(err, &e) && e.Code == 503 {
		return true
	}
	return false
}

func gpuTypesFromConfig() []string {
	var gpuTypeIDs []string
	provider := config.Config.Project.DefaultProvider
	if config.Provider != "" {
		provider = config.Provider
	}
	if p, ok := config.Config.Project.Providers[provider]; ok {
		gpuTypeIDs = p.NodeTypes
	}
	return gpuTypeIDs
}

func renderSessionCreated(exec *types.Exec) {
	if exec == nil {
		return
	}

	results := []ui.ResultEntry{
		{Key: "Name", Value: exec.Name},
		{Key: "ID", Value: exec.ID},
		{Key: "Provider", Value: exec.Provider.DisplayName()},
		{Key: "Instance Type", Value: exec.NodeTypeID},
		{Key: "Region", Value: exec.Region},
		{Key: "Status", Value: fmt.Sprintf("%s", exec.Status)},
		{Key: "SSHKey", Value: fmt.Sprintf("%s", exec.SSHKey.Name)},
		{Key: "CPUs", Value: fmt.Sprintf("%v", exec.Specs.CPU.Min)},
		{Key: "RAM", Value: fmt.Sprintf("%v", exec.Specs.RAM.Min)},
		{Key: "HDD", Value: fmt.Sprintf("%v", exec.Specs.HDD.Min)},
		{Key: "GPU Type", Value: fmt.Sprintf("%v", exec.Specs.GPU.Type)},
		{Key: "NumGPUs", Value: fmt.Sprintf("%v", exec.Specs.GPU.Count.Min)},
	}

	ui.ResultTitle("Session Created:")
	ui.Result(results, ui.IndentWidth)
	return
}
