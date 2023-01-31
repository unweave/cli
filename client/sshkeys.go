package client

import (
	"context"
	"fmt"

	"github.com/unweave/unweave/api/types"
)

type SSHKeyService struct {
	client *Client
}

func (s *SSHKeyService) Add(ctx context.Context, params types.SSHKeyAddRequestParams) error {
	uri := fmt.Sprintf("ssh-keys")
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, params)
	if err != nil {
		return err
	}
	res := &types.SSHKeyAddResponse{}
	if err = s.client.ExecuteRest(ctx, req, res); err != nil {
		return err
	}
	return nil
}

func (s *SSHKeyService) Generate(ctx context.Context, params types.SSHKeyGenerateRequestParams) (*types.SSHKeyGenerateResponse, error) {
	uri := fmt.Sprintf("ssh-keys/generate")
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, params)
	if err != nil {
		return nil, err
	}
	res := &types.SSHKeyGenerateResponse{}
	if err = s.client.ExecuteRest(ctx, req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *SSHKeyService) List(ctx context.Context) ([]types.SSHKey, error) {
	uri := fmt.Sprintf("ssh-keys")
	req, err := s.client.NewAuthorizedRestRequest(Get, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	res := &types.SSHKeyListResponse{}
	if err = s.client.ExecuteRest(ctx, req, res); err != nil {
		return nil, err
	}
	return res.Keys, nil
}
