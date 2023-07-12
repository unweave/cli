package client

import (
	"context"
	"fmt"

	"github.com/unweave/unweave/api/types"
)

type EvalService struct {
	client *Client
}

func (s *EvalService) List(ctx context.Context, userID, projectID string) ([]types.Eval, error) {
	uri := fmt.Sprintf("projects/%s/%s/evals", userID, projectID)
	req, err := s.client.NewAuthorizedRestRequest(Get, uri, nil, nil)
	if err != nil {
		return nil, err
	}

	evals := types.EvalList{}
	if err = s.client.ExecuteRest(ctx, req, &evals); err != nil {
		return nil, err
	}

	return evals.Evals, nil
}

func (s *EvalService) Create(ctx context.Context, userID, projectID, execID string) (types.Eval, error) {
	request := types.EvalCreate{ExecID: execID}

	uri := fmt.Sprintf("projects/%s/%s/evals", userID, projectID)
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, request)
	if err != nil {
		return types.Eval{}, err
	}

	eval := types.Eval{}
	if err = s.client.ExecuteRest(ctx, req, &eval); err != nil {
		return types.Eval{}, err
	}

	return eval, nil
}
