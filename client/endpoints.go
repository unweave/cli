package client

import (
	"context"
	"fmt"

	"github.com/unweave/unweave/api/types"
)

type EndpointService struct {
	client *Client
}

func (s *EndpointService) List(ctx context.Context, userID, projectID string) ([]types.Endpoint, error) {
	uri := fmt.Sprintf("projects/%s/%s/endpoints", userID, projectID)
	req, err := s.client.NewAuthorizedRestRequest(Get, uri, nil, nil)
	if err != nil {
		return nil, err
	}

	endpoints := types.EndpointList{}
	if err = s.client.ExecuteRest(ctx, req, &endpoints); err != nil {
		return nil, err
	}

	return endpoints.Endpoints, nil
}

func (s *EndpointService) Create(ctx context.Context, userID, projectID, execID string) (types.Endpoint, error) {
	request := types.EndpointCreate{ExecID: execID}

	uri := fmt.Sprintf("projects/%s/%s/endpoints", userID, projectID)
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, request)
	if err != nil {
		return types.Endpoint{}, err
	}

	endpoint := types.Endpoint{}
	if err = s.client.ExecuteRest(ctx, req, &endpoint); err != nil {
		return types.Endpoint{}, err
	}

	return endpoint, nil
}

func (s *EndpointService) EvalAttach(ctx context.Context, userID, projectID, endpointID, evalID string) error {
	uri := fmt.Sprintf("projects/%s/%s/endpoints/%s/eval", userID, projectID, endpointID)
	attach := types.EndpointEvalAttach{EvalID: evalID}

	req, err := s.client.NewAuthorizedRestRequest(Put, uri, nil, attach)
	if err != nil {
		return err
	}

	if err = s.client.ExecuteRest(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

func (s *EndpointService) RunEvalCheck(ctx context.Context, userID, projectID, endpointID string) error {
	uri := fmt.Sprintf("projects/%s/%s/endpoints/%s/check", userID, projectID, endpointID)
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, nil)
	if err != nil {
		return err
	}

	if err = s.client.ExecuteRest(ctx, req, nil); err != nil {
		return err
	}

	return nil
}
