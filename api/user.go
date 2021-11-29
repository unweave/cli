package api

import "context"

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
		GeneratePairingCode struct {
			Code string `json:"code"`
		} `json:"generatePairingCode"`
	}
	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return "", err
	}

	return resp.GeneratePairingCode.Code, nil
}

// ExchangePairingCode exchanges a pairing code for an access token.
func (a *Api) ExchangePairingCode(ctx context.Context, code string) (string, error) {
	vars := struct {
		Code string `json:"code"`
	}{
		Code: code,
	}
	req, err := a.NewGqlRequest(`
		mutation ExchangePairingCode ($code: String!) { 
			exchangePairingCode(code: $code){
				token
			}
		}`, vars)
	if err != nil {
		return "", err
	}

	var resp struct {
		Token string `json:"exchangePairingCode"`
	}

	if err = a.ExecuteGql(ctx, req, &resp); err != nil {
		return "", err
	}
	return resp.Token, nil
}
