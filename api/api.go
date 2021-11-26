package api

import "net/http"

type Api struct {
	httpClient *http.Client
}

func New() *Api {
	return &Api{
		httpClient: &http.Client{
			Timeout: 60,
		},
	}
}
