package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

// Create attempts to create a session using the Exec spec provided, uses GPUs in the config if not, returns a 503 out-of-capacity error.
// Renders newly created sessions to the UI implicitly.
func Create(ctx context.Context, params types.ExecCreateParams) (string, error) {
	if params.Spec.GPU.Type == "" {
		exec, err := createSessionFromConfigGPUTypes(ctx, params)
		renderSessionCreated(exec)

		if err != nil {
			return "", err
		}
		return exec.ID, nil
	}

	exec, err := createSession(ctx, params, params.Spec.GPU.Type)
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
	useParams.Spec.GPU.Type = gpuType

	exec, err := uwc.Exec.Create(ctx, owner, projectName, useParams)
	if err != nil {
		return nil, err
	}

	return exec, nil
}

func createSessionFromConfigGPUTypes(ctx context.Context, params types.ExecCreateParams) (*types.Exec, error) {
	gpuTypesFromConfig := gpuTypesFromConfig()

	var exec *types.Exec
	var err error
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

// gpuTypesFromConfig returns the GPU types in config.toml or a set of defaults, never nil
func gpuTypesFromConfig() []string {
	var gpuTypeIDs []string
	provider := config.Config.Project.DefaultProvider
	if config.Provider != "" {
		provider = config.Provider
	}
	if p, ok := config.Config.Project.Providers[provider]; ok {
		gpuTypeIDs = p.NodeTypes
	}
	if len(gpuTypeIDs) == 0 {
		gpuTypeIDs = config.DefaultGPUTypes
	}
	if len(gpuTypeIDs) == 0 {
		ui.HandleError(fmt.Errorf("❌ Please specify default GPU types in .unweave/config.toml and try again"))
		os.Exit(1)
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
		{Key: "Instance Type", Value: exec.ID},
		{Key: "Region", Value: exec.Region},
		{Key: "Status", Value: fmt.Sprintf("%s", exec.Status)},
		{Key: "SSHKeys", Value: fmt.Sprintf("%s", getSSHKeyNames(exec.Keys))},
		{Key: "CPUs", Value: fmt.Sprintf("%v", exec.Spec.CPU.Min)},
		// Uncomment when issues setting RAM are resolved
		// {Key: "RAM", Value: fmt.Sprintf("%vGB", exec.Specs.RAM.Min)},
		{Key: "HDD", Value: fmt.Sprintf("%vGB", exec.Spec.HDD.Min)},
		{Key: "GPU Type", Value: fmt.Sprintf("%s", exec.Spec.GPU.Type)},
		{Key: "NumGPUs", Value: fmt.Sprintf("%v", exec.Spec.GPU.Count.Min)},
		{Key: "Volumes", Value: ui.FormatVolumes(exec.Volumes)},
	}

	if exec.Network.HTTPService != nil {
		results = append(results,
			ui.ResultEntry{Key: "InternalPort", Value: fmt.Sprintf("%d", exec.Network.HTTPService.InternalPort)},
		)
	}

	ui.ResultTitle("Session Created:")
	ui.Result(results, ui.IndentWidth)
	return
}

func getSSHKeyNames(keys []types.SSHKey) string {
	keyNames := make([]string, 0, len(keys))

	for _, key := range keys {
		keyNames = append(keyNames, key.Name)
	}

	return strings.Join(keyNames, ", ")
}
