package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/unweave/unweave/api/types"
)

type ExecService struct {
	client *Client
}

func (s *ExecService) Create(ctx context.Context, owner, project string, params types.ExecCreateParams) (*types.Exec, error) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	if params.Source != nil && params.Source.Context != nil {
		fw, err := w.CreateFormFile("context", "context.zip")
		if err != nil {
			return nil, err
		}
		if _, err = io.Copy(fw, params.Source.Context); err != nil {
			return nil, err
		}
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	if err = w.WriteField("params", string(paramsJSON)); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("projects/%s/%s/sessions", owner, project)

	req, err := http.NewRequest("POST", uri, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+s.client.cfg.Token)

	// TODO: hack for now: add box query if PersistentFS is set
	query := ""
	r := &RestRequest{
		Url:    fmt.Sprintf("%s/%s?%s", s.client.cfg.ApiURL, uri, query),
		Header: req.Header,
		Body:   buf,
		Type:   Post,
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	exec := &types.Exec{}
	if err = s.client.ExecuteRest(ctx, r, exec); err != nil {
		return nil, err
	}
	return exec, nil
}

func (s *ExecService) Exec(ctx context.Context, cmd []string, image string, sessionID *string) (*types.Exec, error) {
	return nil, nil
}

func (s *ExecService) Get(ctx context.Context, owner, project, sessionID string) (*types.Exec, error) {
	uri := fmt.Sprintf("projects/%s/%s/sessions/%s", owner, project, sessionID)
	req, err := s.client.NewAuthorizedRestRequest(Get, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	session := &types.Exec{}
	if err = s.client.ExecuteRest(ctx, req, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *ExecService) List(ctx context.Context, owner, project string, listTerminated bool) ([]types.Exec, error) {
	uri := fmt.Sprintf("projects/%s/%s/sessions", owner, project)
	query := map[string]string{
		"terminated": fmt.Sprintf("%t", listTerminated),
	}
	req, err := s.client.NewAuthorizedRestRequest(Get, uri, query, nil)
	if err != nil {
		return nil, err
	}
	res := &types.SessionsListResponse{}
	if err = s.client.ExecuteRest(ctx, req, res); err != nil {
		return nil, err
	}
	return res.Sessions, nil
}

func (s *ExecService) Terminate(ctx context.Context, owner, project, sessionID string) error {
	uri := fmt.Sprintf("projects/%s/%s/sessions/%s/terminate", owner, project, sessionID)
	req, err := s.client.NewAuthorizedRestRequest(Put, uri, nil, nil)
	if err != nil {
		return err
	}
	res := &types.SessionTerminateResponse{}
	return s.client.ExecuteRest(ctx, req, res)
}
