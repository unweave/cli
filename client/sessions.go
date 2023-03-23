package client

import (
	"context"
	"fmt"

	"github.com/unweave/unweave/api/types"
)

type SessionService struct {
	client *Client
}

func (s *SessionService) Create(ctx context.Context, owner, project string, params types.SessionCreateParams) (*types.Session, error) {
	uri := fmt.Sprintf("projects/%s/%s/sessions", owner, project)
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, params)
	if err != nil {
		return nil, err
	}
	session := &types.Session{}
	if err = s.client.ExecuteRest(ctx, req, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionService) Exec(ctx context.Context, cmd []string, image string, sessionID *string) (*types.Session, error) {
	return nil, nil
}

func (s *SessionService) Get(ctx context.Context, owner, project, sessionID string) (*types.Session, error) {
	uri := fmt.Sprintf("projects/%s/%s/sessions/%s", owner, project, sessionID)
	req, err := s.client.NewAuthorizedRestRequest(Get, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	session := &types.Session{}
	if err = s.client.ExecuteRest(ctx, req, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SessionService) List(ctx context.Context, owner, project string, listTerminated bool) ([]types.Session, error) {
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

func (s *SessionService) Terminate(ctx context.Context, owner, project, sessionID string) error {
	uri := fmt.Sprintf("projects/%s/%s/sessions/%s/terminate", owner, project, sessionID)
	req, err := s.client.NewAuthorizedRestRequest(Put, uri, nil, nil)
	if err != nil {
		return err
	}
	res := &types.SessionTerminateResponse{}
	return s.client.ExecuteRest(ctx, req, res)
}
