package api

import (
	"context"
	"github.com/unweave/cli/entity"
)

// GetUser returns a user by id
func (a *Api) GetUser(ctx context.Context, id int64, email string) (*entity.User, error) {
	vars := struct {
		Id    int64  `json:"id"`
		Email string `json:"email"`
	}{
		Id:    id,
		Email: email,
	}
	req, err := a.NewGqlRequest(`
		query GetUser ($id: BigInt, $email: String) {
			user (email: $email, id: $id) {
				id
				email
			}
		}`, vars)

	if err != nil {
		return nil, err
	}

	var resp struct {
		User entity.User `json:"me"`
	}

	err = a.ExecuteGql(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.User, nil
}

// GeneratePairingCode generates a new auth code for a user to pair their CLI to their
// account through the webapp.
func (a *Api) GeneratePairingCode(ctx context.Context) (string, error) {
	req, err := a.NewGqlRequest(`
		mutation { 
			generatePairingCode {
				code
			}
		 }`, struct{}{})
	if err != nil {
		return "", err
	}

	var resp struct {
		Data entity.GeneratePairingCode `json:"generatePairingCode"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return "", err
	}

	return resp.Data.Code, nil
}

// ExchangePairingCode exchanges a pairing code for an access token.
func (a *Api) ExchangePairingCode(ctx context.Context, code string) (string, string, error) {
	vars := struct {
		Code string `json:"code"`
	}{
		Code: code,
	}
	req, err := a.NewGqlRequest(`
		mutation ExchangePairingCode ($code: String!) { 
			exchangePairingCode(code: $code){
				uid
				token
			}
		}`, vars)
	if err != nil {
		return "", "", err
	}

	var resp struct {
		Data entity.ExchangePairingCode `json:"exchangePairingCode"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return "", "", err
	}
	return resp.Data.Uid, resp.Data.Token, nil
}
