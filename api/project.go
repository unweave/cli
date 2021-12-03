package api

import (
	"context"
	"github.com/unweave/cli/entity"
)

func (a *Api) GetUserProject(ctx context.Context, projectId string) (*entity.Project, error) {
	req, err := a.NewAuthorizedGqlRequest(entity.GetProjectQuery, struct {
		Id string `json:"id"`
	}{
		Id: projectId,
	})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data entity.Project `json:"project"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (a *Api) GetUserProjects(ctx context.Context) ([]entity.Project, error) {
	req, err := a.NewAuthorizedGqlRequest(entity.GetProjectsQuery, struct{}{})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []entity.Project `json:"projects"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
