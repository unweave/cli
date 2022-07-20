package api

import (
	"context"

	"github.com/unweave/cli/model"
)

func (a *Api) CreateUserToken(ctx context.Context) (*model.UserToken, error) {
	req, err := a.NewAuthorizedGqlRequest(model.CreateUserTokenMutation, struct{}{})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data model.UserToken `json:"createUserToken"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (a *Api) GetUserTokens(ctx context.Context) ([]*model.UserToken, error) {
	req, err := a.NewAuthorizedGqlRequest(model.GetUserTokensQuery, struct{}{})
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []*model.UserToken `json:"userTokens"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
