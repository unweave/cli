package session

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/unweave/cli/client"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

// Create attempts to create a session using the Exec spec provided, uses GPUs in the config if not, returns a 503 out-of-capacity error.
// Renders newly created sessions to the UI implicitly.
func Create(ctx context.Context, params types.ExecCreateParams) (string, error) {
	factory := newSessionFactory()

	exec, err := factory.createSession(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error creating session: %w", err)
	}

	renderSessionCreated(exec)

	return exec.ID, nil
}

type sessionFactory struct {
	uwc         *client.Client
	projectName string
	owner       string
}

func newSessionFactory() sessionFactory {
	uwc := config.InitUnweaveClient()
	owner, projectName := config.GetProjectOwnerAndName()

	return sessionFactory{uwc: uwc, projectName: projectName, owner: owner}
}

func (s sessionFactory) createSession(ctx context.Context, params types.ExecCreateParams) (*types.Exec, error) {
	var (
		exec *types.Exec
		err  error

		hasGpuType = params.Spec.GPU.Type != ""
		hasCpuType = params.Spec.CPU.Type != ""
	)

	switch {
	case hasGpuType && hasCpuType:
		// error, not possible to fulfil
		return nil, &types.Error{
			Message:    "Cannot set gpu type and cpu type",
			Suggestion: "Set just one of gpu type and cpu type",
			Err:        errors.New("both gpu and cpu types set"),
		}

	case !hasGpuType && !hasCpuType:
		exec, err = s.createSessionFromConfigNodeTypes(ctx, params)

	case hasGpuType:
		exec, err = s.createExecSession(ctx, params)

	case hasCpuType:
		exec, err = s.createExecSession(ctx, params)
	}
	if err != nil {
		return nil, err
	}

	return exec, nil
}

func (s sessionFactory) createExecSession(ctx context.Context, params types.ExecCreateParams) (*types.Exec, error) {
	exec, err := s.uwc.Exec.Create(ctx, s.owner, s.projectName, params)
	if err != nil {
		return nil, err
	}

	return exec, nil
}

func (s sessionFactory) createSessionFromConfigNodeTypes(ctx context.Context, params types.ExecCreateParams) (*types.Exec, error) {
	nodeTypesFromConfig := nodeTypesFromConfig()

	const includeOnlyAvailable = true

	nodeTypes, err := s.uwc.Provider.ListNodeTypes(ctx, params.Provider, includeOnlyAvailable)
	if err != nil {
		return nil, &types.Error{
			Message:    "Cannot find gpu and cpu node types",
			Suggestion: "",
			Err:        errors.New("failed to list node types"),
		}
	}

	gpuNodeType := make(map[string]bool)

	for _, n := range nodeTypes {
		gpuNodeType[n.ID] = n.Type == "GPU"
	}

	for _, nodeType := range nodeTypesFromConfig {
		if gpuNodeType[nodeType] {
			params.Spec.GPU.Type = nodeType
		} else {
			params.Spec.CPU.Type = nodeType
		}

		exec, err := s.createExecSession(ctx, params)
		if err != nil {
			if isOutOfCapacityError(err) {
				continue
			}
			return nil, err
		}

		return exec, nil
	}

	return nil, &types.Error{
		Message:    "No default node type provided",
		Suggestion: "Update node_types in .unweave/config.toml or pass --gpu-type flag",
		Err:        errors.New("no default node types"),
	}
}

func isOutOfCapacityError(err error) bool {
	var e *types.Error
	if errors.As(err, &e) && e.Code == 503 {
		return true
	}
	return false
}

// nodeTypesFromConfig returns the GPU types in config.toml or a set of defaults, never nil
func nodeTypesFromConfig() []string {
	var nodeTypeIDs []string
	provider := config.Config.Project.DefaultProvider
	if config.Provider != "" {
		provider = config.Provider
	}
	if p, ok := config.Config.Project.Providers[provider]; ok {
		nodeTypeIDs = p.NodeTypes
	}
	if len(nodeTypeIDs) == 0 {
		nodeTypeIDs = config.DefaultNodeTypes
	}

	return nodeTypeIDs
}

func renderSessionCreated(exec *types.Exec) {
	if exec == nil {
		return
	}

	nodeType := exec.Spec.GPU.Type
	if nodeType == "" {
		nodeType = exec.Spec.CPU.Type
	}

	results := []ui.ResultEntry{
		{Key: "Name", Value: exec.Name},
		{Key: "ID", Value: exec.ID},
		{Key: "Provider", Value: exec.Provider.DisplayName()},
		{Key: "Region", Value: exec.Region},
		{Key: "Status", Value: fmt.Sprintf("%s", exec.Status)},
		{Key: "SSHKeys", Value: fmt.Sprintf("%s", getSSHKeyNames(exec.Keys))},
		{Key: "CPUs", Value: fmt.Sprintf("%v", exec.Spec.CPU.Min)},
		// Uncomment when issues setting RAM are resolved
		// {Key: "RAM", Value: fmt.Sprintf("%vGB", exec.Specs.RAM.Min)},
		{Key: "HDD", Value: fmt.Sprintf("%vGB", exec.Spec.HDD.Min)},
		{Key: "Node type", Value: nodeType},
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
