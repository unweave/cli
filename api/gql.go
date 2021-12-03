package api

import (
	"context"
	"fmt"
	"github.com/unweave/cli/pkg/graphql"
	"reflect"
)

// NewGqlRequest creates a new request for the GraphQL API and attaches the given
// variables to the query.
func (a *Api) NewGqlRequest(query string, vars interface{}) (*graphql.Request, error) {
	req := graphql.NewRequest(query)
	fields := reflect.ValueOf(vars)
	for i := 0; i < fields.NumField(); i++ {
		k := fields.Type().Field(i).Tag.Get("json")
		v := fields.Field(i).Interface()

		req.Var(k, v)
	}

	return req, nil

}

// NewAuthorizedGqlRequest extends NewGqlRequest by attaching authorization headers from
// the user's config.
func (a *Api) NewAuthorizedGqlRequest(query string, vars interface{}) (*graphql.Request, error) {
	req, err := a.NewGqlRequest(query, vars)
	if err != nil {
		return nil, err
	}
	if a.cfg.Root.User == nil || a.cfg.Root.User.Token == "" {
		fmt.Println("You are not logged in. Please run `unweave login` to login.")
		return nil, fmt.Errorf("not logged in")
	}

	req.Header.Set("Authorization", "Bearer "+a.cfg.Root.User.Token)

	return req, nil
}

// ExecuteGql executes a graphql request and populates the resp interface with the result.
// The resp interface must be a pointer to a struct.
func (a *Api) ExecuteGql(ctx context.Context, req *graphql.Request, resp interface{}) error {
	if err := a.gql.Run(ctx, req, resp); err != nil {
		return err
	}

	return nil
}
