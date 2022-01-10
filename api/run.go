package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/unweave/cli/entity"
	"log"
	"mime/multipart"
)

func (a *Api) CreateRunSession(ctx context.Context, projectId string) (string, error) {
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

func (a *Api) UploadRunContext(ctx context.Context, projectId, runId string, gatherContext entity.GatherContextFunc) error {
	endpoint := fmt.Sprintf("api/project/%s/run-session/%s/upload-context", projectId, runId)
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

func (a *Api) GetRunStatus(ctx context.Context, runId string) (string, error) {
	return "", nil
}

func (a *Api) ConnectToZepl(ctx context.Context, projectId, runId string) error {
	endpoint := fmt.Sprintf("api/project/%s/run-session/%s/follow", projectId, runId)

	done, conn, err := a.NewSocketConnection(ctx, endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()
	defer close(done)

	fmt.Println("Connected to Zepl")
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
	}

	fmt.Println("Disconnected from Zepl")
	return nil
}
