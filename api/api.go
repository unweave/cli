package api

import (
	"context"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/pkg/graphql"
	"log"
	"net/http"
	"time"
)

type Api struct {
	cfg  *config.Config
	gql  *graphql.Client
	rest *http.Client
}

type Execute func(ctx context.Context, resp interface{}) error

func GetApiUrl() string {
	return "http://localhost:4000"
}

func GetAppUrl() string {
	return "http://localhost:3000"
}

func GetGqlUrl() string {
	return GetApiUrl() + "/"
}

func GetRestUrl() string {
	return GetAppUrl() + "/api"
}

func New() *Api {
	cfg := config.New()
	gqlClient := graphql.NewClient(GetGqlUrl())
	httpClient := &http.Client{Timeout: time.Second * 60}

	if cfg.IsDebug {
		gqlClient.Log = func(s string) { log.Println(s) }
	}
	return &Api{
		cfg:  config.New(),
		gql:  gqlClient,
		rest: httpClient,
	}
}
