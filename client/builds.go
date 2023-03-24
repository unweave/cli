package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/unweave/unweave/api/types"
)

type BuildService struct {
	client *Client
}

func (b *BuildService) Create(ctx context.Context, owner, project string, params types.BuildsCreateParams) (string, error) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	fw, err := w.CreateFormFile("context", "context.zip")
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, params.BuildContext); err != nil {
		return "", err
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	if err = w.WriteField("params", string(paramsJSON)); err != nil {
		return "", err
	}
	if err = w.Close(); err != nil {
		return "", err
	}

	uri := fmt.Sprintf("projects/%s/%s/builds", owner, project)
	req, err := http.NewRequest("POST", uri, buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+b.client.cfg.Token)

	r := &RestRequest{
		Url:    fmt.Sprintf("%s/%s?%s", b.client.cfg.ApiURL, uri, ""),
		Header: req.Header,
		Body:   buf,
		Type:   Post,
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res := &types.BuildsCreateResponse{}
	if err = b.client.ExecuteRest(ctx, r, res); err != nil {
		return "", err
	}
	return res.BuildID, nil
}
