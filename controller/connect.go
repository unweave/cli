package controller

import "context"

func (c *Controller) Connect(ctx context.Context, projectId, zeplId string) error {
	// check if project ID and run ID exist
	// check if run session is active
	// connect to run session

	// Connect to get logs
	if err := c.api.ConnectToZepl(ctx, projectId, zeplId); err != nil {
		return err
	}
	return nil
}
