package controller

import (
	"context"
)

func (c *Controller) Logs(ctx context.Context, zeplID string) error {
	if err := c.api.TailZeplLogs(ctx, zeplID); err != nil {
		return err
	}
	return nil
}
