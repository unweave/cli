package project

import (
	"context"
	"fmt"
	"github.com/unweave/cli/config"
	"github.com/unweave/unweave/api/types"
)

func Create(ctx context.Context, name, visibility string) (types.Project, error) {
	client := config.InitUnweaveClient()
	owner := config.GetActiveAccountID()

	params := types.ProjectCreateRequestParams{
		Name:          name,
		Tags:          nil,
		Visibility:    nil,
		SourceRepoURL: nil,
	}
	project, err := client.Account.ProjectCreate(ctx, owner, params)
	if err != nil {
		return types.Project{}, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}
