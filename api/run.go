package api

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/unweave/cli/model"
)

func computeContentHash(r io.Reader) (string, error) {
	hasher := sha1.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	sha := hex.EncodeToString(hasher.Sum(nil))
	return sha, nil
}

func (a *Api) CreateZepl(ctx context.Context, projectID, command string) (
	*model.Zepl, error,
) {
	req, err := a.NewAuthorizedGqlRequest(model.InitZeplMutation, struct {
		ProjectID string `json:"projectID"`
		Command   string `json:"command"`
		Gpu       bool   `json:"gpu"`
	}{
		ProjectID: projectID,
		Command:   command,
		Gpu:       a.cfg.Zepl.IsGpu,
	})

	if err != nil {
		return nil, err
	}

	var resp struct {
		Data model.Zepl `json:"initZepl"`
	}
	if err := a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (a *Api) UploadZeplContext(ctx context.Context, zeplID string, gatherContext model.GatherContextFunc) error {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("context", "context.zip")

	// Create the context to be uploaded
	if err = gatherContext(part); err != nil {
		return err
	}
	writer.Close()

	endpoint := fmt.Sprintf("zepl/%s/upload", zeplID)
	req, err := a.NewAuthorizedRestRequest(Post, endpoint, nil)
	if err != nil {
		return err
	}

	req.Body = buf
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var resp interface{}
	if err = a.ExecuteRest(ctx, req, &resp); err != nil {
		return err
	}
	return nil
}

func (a *Api) LaunchZepl(ctx context.Context, zeplID string) error {
	endpoint := fmt.Sprintf("zepl/%s/launch", zeplID)

	req, err := a.NewAuthorizedRestRequest(Post, endpoint, nil)
	if err != nil {
		return err
	}

	var resp interface{}
	if err = a.ExecuteRest(ctx, req, &resp); err != nil {
		return err
	}
	return nil
}

// TailZeplLogs prints logs for a zepl
// TODO: reimplement to return a channel that will receive logs from the Zepl
func (a *Api) TailZeplLogs(ctx context.Context, zeplID string) error {
	endpoint := fmt.Sprintf("zepl/%s/logs", zeplID)

	done, conn, err := a.NewSocketConnection(ctx, endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()
	defer close(done)

	fmt.Printf("Connected to Zepl with ID %s\n", zeplID)
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
