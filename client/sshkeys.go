package client

import (
	"context"
	"fmt"

	"github.com/unweave/unweave/api/types"
)

type SSHKeyService struct {
	client *Client
}

func (s *SSHKeyService) Add(ctx context.Context, owner string, params types.SSHKeyAddParams) (string, error) {
	uri := fmt.Sprintf("ssh-keys/%s", owner)
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, params)
	if err != nil {
		return "", err
	}
	res := &types.SSHKeyResponse{}
	if err = s.client.ExecuteRest(ctx, req, res); err != nil {
		return "", err
	}
	return res.Name, nil
}

func (s *SSHKeyService) Generate(ctx context.Context, owner string, params types.SSHKeyGenerateParams) (*types.SSHKeyResponse, error) {
	uri := fmt.Sprintf("ssh-keys/%s/generate", owner)
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, params)
	if err != nil {
		return nil, err
	}
	res := &types.SSHKeyResponse{}
	if err = s.client.ExecuteRest(ctx, req, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *SSHKeyService) List(ctx context.Context, owner string) ([]types.SSHKey, error) {
	uri := fmt.Sprintf("ssh-keys/%s", owner)
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
