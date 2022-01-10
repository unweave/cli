package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/unweave/cli/entity"
	"mime/multipart"
)

func (a *Api) CreateZepl(ctx context.Context, projectId string) (string, error) {
	endpoint := fmt.Sprintf("api/project/%s/run-session", projectId)
	req, err := a.NewAuthorizedRestRequest(Post, endpoint, nil)
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

func (a *Api) UploadZeplContext(ctx context.Context, projectId, zeplId string, gatherContext entity.GatherContextFunc) error {
	endpoint := fmt.Sprintf("api/project/%s/run-session/%s/upload-context", projectId, zeplId)
	req, err := a.NewAuthorizedRestRequest(Post, endpoint, nil)
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
	if err = a.ExecuteRest(ctx, req, &resp); err != nil {
		return err
	}
	return nil
}

func (a *Api) GetRunStatus(ctx context.Context, zeplId string) (string, error) {
	return "", nil
}

func (a *Api) ConnectToZepl(ctx context.Context, projectId, zeplId string) error {
	endpoint := fmt.Sprintf("api/project/%s/run-session/%s/follow", projectId, zeplId)

	done, conn, err := a.NewSocketConnection(ctx, endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()
	defer close(done)

	fmt.Printf("Connected to Zepl with ID %s\n", zeplId)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		fmt.Printf("%s\n", message)
	}

	fmt.Println("Disconnected from Zepl")
	return nil
}
