package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/pkg/graphql"
)

type Api struct {
	cfg  *config.Config
	gql  *graphql.Client
	rest *http.Client
}

type Execute func(ctx context.Context, resp interface{}) error

func New() *Api {
	cfg := config.New()
	gqlClient := graphql.NewClient(cfg.Api.GqlUrl)
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
