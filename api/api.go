package api

import (
	"context"
	"github.com/unweave/cli/config"
	"github.com/unweave/cli/pkg/graphql"
	"log"
)

type Api struct {
	cfg *config.Config
	gql *graphql.Client
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

func New() *Api {
	cfg := config.New()
	client := graphql.NewClient(GetGqlUrl())
	if cfg.IsDebug {
		client.Log = func(s string) { log.Println(s) }
	}
	return &Api{
		gql: client,
		cfg: config.New(),
	}
}
