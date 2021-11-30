package api

import "context"

func (a *Api) CreateRunSession(ctx context.Context) error {
	return nil
}

func (a *Api) UploadRunContext(ctx context.Context, runId string) error {
	return nil
}

func (a *Api) GetRunStatus(ctx context.Context, runId string) (string, error) {
	return "", nil
}
