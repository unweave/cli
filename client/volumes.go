package client

import (
	"context"
	"fmt"

	"github.com/unweave/unweave/api/types"
)

type VolumeService struct {
	client *Client
}

func (s *VolumeService) Create(ctx context.Context, userID, projectID string, create types.VolumeCreateRequest) (types.Volume, error) {
	uri := fmt.Sprintf("projects/%s/%s/volumes", userID, projectID)
	req, err := s.client.NewAuthorizedRestRequest(Post, uri, nil, create)
	if err != nil {
		return types.Volume{}, err
	}

	vol := &types.Volume{}
	if err = s.client.ExecuteRest(ctx, req, vol); err != nil {
		return types.Volume{}, err
	}

	return *vol, nil
}

func (s *VolumeService) Delete(ctx context.Context, userID, projectID, volumeIDOrName string) error {
	uri := fmt.Sprintf("projects/%s/%s/volumes/%s", userID, projectID, volumeIDOrName)
	req, err := s.client.NewAuthorizedRestRequest(Delete, uri, nil, nil)
	if err != nil {
		return err
	}

	if err = s.client.ExecuteRest(ctx, req, nil); err != nil {
		return err
	}

	return nil
}

func (s *VolumeService) List(ctx context.Context, userID string, projectID string) ([]types.Volume, error) {
	uri := fmt.Sprintf("projects/%s/%s/volumes", userID, projectID)
	req, err := s.client.NewAuthorizedRestRequest(Get, uri, nil, nil)
	if err != nil {
		return nil, err
	}
	if err = s.client.ExecuteRest(ctx, req, nil); err != nil {
		return nil, err
	}

	res := &types.VolumesListResponse{}
	if err = s.client.ExecuteRest(ctx, req, res); err != nil {
		return nil, err
	}

	return res.Volumes, err
}

func (s *VolumeService) Update(ctx context.Context, userID, projectID, volumeIDOrName string, update types.VolumeResizeRequest) error {
	uri := fmt.Sprintf("projects/%s/%s/volumes/%s", userID, projectID, volumeIDOrName)
	req, err := s.client.NewAuthorizedRestRequest(Put, uri, nil, update)
	if err != nil {
		return err
	}

	if err = s.client.ExecuteRest(ctx, req, nil); err != nil {
		return err
	}

	return nil
}
