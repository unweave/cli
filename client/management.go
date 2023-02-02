package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/unweave/unweave/api/types"
)

type AccountService struct {
	client *Client
}

func (a *AccountService) AccountGet(ctx context.Context, accountID uuid.UUID) (types.Account, error) {
	uri := fmt.Sprintf("account/%s", accountID)
	req, err := a.client.NewAuthorizedRestRequest(Get, uri, nil, nil)
	if err != nil {
		return types.Account{}, err
	}
	res := &types.AccountGetResponse{}
	if err = a.client.ExecuteRest(ctx, req, res); err != nil {
		return types.Account{}, err
	}
	return res.Account, nil
}

func (a *AccountService) PairingTokenCreate(ctx context.Context) (code string, err error) {
	uri := fmt.Sprintf("account/pair")
	req, err := a.client.NewAuthorizedRestRequest(Post, uri, nil, nil)
	if err != nil {
		return "", err
	}
	res := &types.PairingTokenCreateResponse{}
	if err = a.client.ExecuteRest(ctx, req, res); err != nil {
		return "", err
	}
	return res.Code, nil
}

func (a *AccountService) PairingTokenExchange(ctx context.Context, code string) (token string, account *types.Account, err error) {
	uri := fmt.Sprintf("account/pair/%s", code)
	req, err := a.client.NewAuthorizedRestRequest(Put, uri, nil, nil)
	if err != nil {
		return "", nil, err
	}
	res := &types.PairingTokenExchangeResponse{}
	if err = a.client.ExecuteRest(ctx, req, res); err != nil {
		return "", nil, err
	}
	return res.Token, &res.Account, nil
}

func (a *AccountService) ProjectGet(ctx context.Context, projectID uuid.UUID) (types.Project, error) {
	uri := fmt.Sprintf("projects/%s", projectID)
	req, err := a.client.NewAuthorizedRestRequest(Get, uri, nil, nil)
	if err != nil {
		return types.Project{}, err
	}
	res := &types.ProjectGetResponse{}
	if err = a.client.ExecuteRest(ctx, req, res); err != nil {
		return types.Project{}, err
	}
	return res.Project, nil
}
