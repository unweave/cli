package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	goErr "github.com/pkg/errors"
)

type RestRequestType string

const (
	Get    RestRequestType = http.MethodGet
	Post   RestRequestType = http.MethodPost
	Put    RestRequestType = http.MethodPut
	Delete RestRequestType = http.MethodDelete
)

type RestRequest struct {
	Url    string
	Header http.Header
	Body   io.Reader
	Type   RestRequestType
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func (a *Api) NewRestRequest(rtype RestRequestType, endpoint string, params map[string]string) (
	*RestRequest, error,
) {
	query := ""
	fields := reflect.ValueOf(params)
	if params != nil {
		for i := 0; i < fields.NumField(); i++ {
			k := fields.Type().Field(i).Tag.Get("json")
			v := fields.Field(i).Interface()

			query += fmt.Sprintf("%s=%v&", k, v)
		}
	}

	base := a.cfg.Api.WorkbenchUrl
	url := fmt.Sprintf("%s/%s?%s", base, endpoint, query)
	header := http.Header{}
	header.Set("Content-Type", "application/json")

	return &RestRequest{
		Url:    url,
		Header: header,
		Body:   &bytes.Buffer{},
		Type:   rtype,
	}, nil
}

func (a *Api) NewAuthorizedRestRequest(rtype RestRequestType, endpoint string, params map[string]string) (
	*RestRequest, error,
) {
	req, err := a.NewRestRequest(rtype, endpoint, params)
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

func (a *Api) ExecuteRest(ctx context.Context, req *RestRequest, resp interface{}) error {
	httpReq, err := http.NewRequest(string(req.Type), req.Url, req.Body)
	if err != nil {
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	httpReq.Header = req.Header
	res, err := a.rest.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, res.Body); err != nil {
		return goErr.Wrap(err, fmt.Sprintf("status %s, fail to read response body", res.Status))
	}
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		var msg ErrorMessage
		json.NewDecoder(&buf).Decode(&msg)
		return fmt.Errorf("status %s, %s", res.Status, msg.Message)
	}
	if err = json.NewDecoder(&buf).Decode(&resp); err == io.EOF {
		return nil
	} else if err != nil {
		return goErr.Wrap(err, "failed to decode response body")
	}
	return nil
}
