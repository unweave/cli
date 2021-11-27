package api

import (
	"context"
	"github.com/machinebox/graphql"
	"reflect"
)

func (a *Api) NewGqlRequest(query string, vars interface{}) (*graphql.Request, error) {
	req := graphql.NewRequest(query)
	fields := reflect.ValueOf(vars)
	for i := 0; i < fields.NumField(); i++ {
		k := fields.Type().Field(i).Name
		v := fields.Field(i).Interface()

		req.Var(k, v)
	}

	return req, nil

}

func (a *Api) NewAuthorizedGqlRequest(query string, vars interface{}) (*graphql.Request, error) {
	req, err := a.NewGqlRequest(query, vars)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.cfg.Root.User.Token)

	return req, nil
}

func (a *Api) ExecuteGql(ctx context.Context, req *graphql.Request, resp interface{}) error {

	if err := a.gql.Run(ctx, req, &resp); err != nil {
		return err
	}

	return nil
}
