package controller

import "context"

func (c *Controller) Connect(ctx context.Context, projectId, runId string) error {
	// check if project ID and run IS exist
	// check if run session is active
	// connect to run session

	// Connect to get logs
	if err := c.api.ConnectToZepl(ctx, projectId, runId); err != nil {
		return err
	}
	return nil
}
