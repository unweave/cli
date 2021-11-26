package api

import "net/http"

type GraphqlRequest struct {
	query      string
	auth       string
	httpClient http.Client
}

func (a *Api) NewGqlRequest() error {
	return nil
}

func (a *Api) NewAuthorizedGqlRequest() error {
	return nil
}
