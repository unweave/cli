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

var uwc *client.Client

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
	if uwc == nil {
		uwc = config.InitUnweaveClient()
	}

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

	if !hasGpuType && !hasCpuType {
		return nil, &types.Error{
			Message:    "Cannot create an exec with no CPU or GPU type",
			Suggestion: "Set one of cpu or gpu type",
			Provider:   params.Provider,
			Err:        errors.New("missing both cpu and gpu type"),
		}
		//exec, err = s.createSessionFromConfigNodeTypes(ctx, params)
	}

	exec, err = s.createExecSession(ctx, params)
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

func isOutOfCapacityError(err error) bool {
	var e *types.Error
	if errors.As(err, &e) && e.Code == 503 {
		return true
	}
	return false
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
		{Key: "Instance Type", Value: exec.ID},
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
