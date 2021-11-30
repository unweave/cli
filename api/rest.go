package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	goErr "github.com/pkg/errors"
	"io"
	"net/http"
	"reflect"
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
	Body   interface{}
	Type   RestRequestType
}

func (a *Api) NewRestRequest(rtype RestRequestType, endpoint string, params map[string]string) *RestRequest {
	query := ""
	fields := reflect.ValueOf(params)
	for i := 0; i < fields.NumField(); i++ {
		k := fields.Type().Field(i).Tag.Get("json")
		v := fields.Field(i).Interface()

		query += fmt.Sprintf("%s=%v&", k, v)
	}

	base := GetRestUrl()
	url := fmt.Sprintf("%s/%s?%s", base, endpoint, query)
	header := http.Header{}
	header.Set("Content-Type", "application/json")

	return &RestRequest{
		Url:    url,
		Header: header,
		Body:   nil,
		Type:   rtype,
	}
}

func (a *Api) NewAuthorizedRestRequest(rtype RestRequestType, endpoint string, params map[string]string) *RestRequest {
	req := a.NewRestRequest(rtype, endpoint, params)
	req.Header.Set("Authorization", "Bearer "+a.cfg.Root.User.Token)
	return req
}

func (a *Api) ExecuteRest(ctx context.Context, req *RestRequest, resp interface{}) error {
	httpReq, err := http.NewRequest(string(req.Type), req.Url, nil)
	if err != nil {
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	res, err := a.rest.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, res.Body); err != nil {
		return goErr.Wrap(err, fmt.Sprintf("status %s, fail to read response body", res.Status))
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status %s", res.Status)
	}
	if err = json.NewDecoder(&buf).Decode(&resp); err != nil {
		return goErr.Wrap(err, "fail to decode response body")
	}
	return nil
}
