package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/unweave/cli/entity"
	"mime/multipart"
)

func (a *Api) CreateRunSession(ctx context.Context) (string, error) {
	req, err := a.NewAuthorizedRestRequest(Post, "run-session", nil)
	if err != nil {
		return "", err
	}

	var resp struct {
		Id string `json:"id"`
	}
	if err := a.ExecuteRest(ctx, req, &resp); err != nil {
		return "", err
	}
	return resp.Id, nil
}

func (a *Api) UploadRunContext(ctx context.Context, runId string, gatherContext entity.GatherContextFunc) error {
	req, err := a.NewAuthorizedRestRequest(Post, "run-session/"+runId+"/upload-context", nil)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("session_context", "context.zip")

	// Create the context to be uploaded
	if err = gatherContext(part); err != nil {
		return err
	}
	writer.Close()

	req.Body = buf
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
