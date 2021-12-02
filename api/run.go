package api

import (
	"context"
	"fmt"
)

func (a *Api) CreateRunSession(ctx context.Context) (string, error) {
	req := a.NewAuthorizedRestRequest(Post, "run-session", nil)

	var resp struct {
		Id string `json:"id"`
	}
	if err := a.ExecuteRest(ctx, req, &resp); err != nil {
		return "", err
	}

	fmt.Println(resp.Id)
	return "", nil
}

func (a *Api) UploadRunContext(ctx context.Context, runId string) error {
	return nil
}

func (a *Api) GetRunStatus(ctx context.Context, runId string) (string, error) {
	return "", nil
}
