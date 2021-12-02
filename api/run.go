package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
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

func (a *Api) UploadRunContext(ctx context.Context, runId string, buf io.Reader) error {
	req := a.NewAuthorizedRestRequest(Post, "run-session/"+runId+"/upload-context", nil)

	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)
	part, err := writer.CreateFormFile("session_context", "context.zip")
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, buf); err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}

	req.Body = bodyBuffer
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
