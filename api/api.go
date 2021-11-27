package api

import (
	"context"
	"github.com/machinebox/graphql"
	"github.com/unweave/cli/config"
)

type Api struct {
	cfg *config.Config
	gql *graphql.Client
}

type Execute func(ctx context.Context, resp interface{}) error

func GetApiUrl() string {
	return "http://localhost:8080"
}

func GetGqlUrl() string {
	return GetApiUrl() + "/graphql"
}

func New() *Api {
	return &Api{
		gql: graphql.NewClient(GetGqlUrl()),
	}
}
