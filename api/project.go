package api

import (
	"context"

	"github.com/unweave/cli/model"
)

func (a *Api) GetUserProject(ctx context.Context, projectID string) (
	*model.Project, error,
) {
	req, err := a.NewAuthorizedGqlRequest(model.GetProjectQuery, struct {
		ID string `json:"id"`
	}{
		ID: projectID,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data model.Project `json:"project"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (a *Api) GetUserProjects(ctx context.Context) ([]*model.Project, error) {
	req, err := a.NewAuthorizedGqlRequest(model.GetProjectsQuery, struct{}{})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []*model.Project `json:"projects"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
