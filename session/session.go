package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

// Create attempts to create a session using the node types provided
// until the first successful creation. If none of the node types are successful, it
// returns 503 out of capacity error.
func Create(ctx context.Context, params types.ExecCreateParams) (string, error) {
	uwc := config.InitUnweaveClient()

	var err error
	var exec *types.Exec

	// right now all NodeTypes _are_ GPU types
	owner, projectName := config.GetProjectOwnerAndName()
	exec, err = uwc.Exec.Create(ctx, owner, projectName, params)
	if err == nil {
		results := []ui.ResultEntry{
			{Key: "Name", Value: exec.Name},
			{Key: "ID", Value: exec.ID},
			{Key: "Provider", Value: exec.Provider.DisplayName()},
			{Key: "Type", Value: exec.NodeTypeID},
			{Key: "Region", Value: exec.Region},
			{Key: "Status", Value: fmt.Sprintf("%s", exec.Status)},
			{Key: "SSHKey", Value: fmt.Sprintf("%s", exec.SSHKey.Name)},
		}

		ui.ResultTitle("Session Created:")
		ui.Result(results, ui.IndentWidth)
		return exec.ID, nil
	}

	if err != nil {
		var e *types.Error
		if errors.As(err, &e) {
			return "", err
		}
	}
	// Return the last error - which will be a 503 if it's an out of capacity error.
	return "", err
}
