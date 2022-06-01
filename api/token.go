package api

import (
	"context"
	"github.com/unweave/cli/entity"
)

func (a *Api) CreateUserToken(ctx context.Context) (*entity.UserToken, error) {
	req, err := a.NewAuthorizedGqlRequest(entity.CreateUserTokenMutation, struct{}{})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data entity.UserToken `json:"createUserToken"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (a *Api) GetUserTokens(ctx context.Context) ([]*entity.UserToken, error) {
	req, err := a.NewAuthorizedGqlRequest(entity.GetUserTokensQuery, struct{}{})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []*entity.UserToken `json:"userTokens"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
