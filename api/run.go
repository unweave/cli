package api

import (
	"context"
	"fmt"
	"io"
)

func (a *Api) CreateRunSession(ctx context.Context) (string, error) {
	req := a.NewAuthorizedRestRequest(Post, "run-session", nil)

	var resp struct {
		Id string `json:"id"`
	}
	if err := a.ExecuteRest(ctx, req, &resp); err != nil {
		return "", err
	}
	return resp.Id, nil
}

func (a *Api) UploadRunContext(ctx context.Context, runId string, buf io.Reader, headerBoundary string) error {
	req := a.NewAuthorizedRestRequest(Post, "run-session/"+runId+"/upload-context", nil)
	req.Body = buf
	req.Header.Set("Content-Type", headerBoundary)

	var resp interface{}
	if err := a.ExecuteRest(ctx, req, &resp); err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}

func (a *Api) GetRunStatus(ctx context.Context, runId string) (string, error) {
	return "", nil
}
