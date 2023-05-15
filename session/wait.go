package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/unweave/cli/config"
	"github.com/unweave/cli/ui"
	"github.com/unweave/unweave/api/types"
)

func Wait(ctx context.Context, execID string) (execch chan types.Exec, errch chan error, err error) {
	uwc := config.InitUnweaveClient()
	listTerminated := config.All
	owner, projectName := config.GetProjectOwnerAndName()

	errch = make(chan error)
	execch = make(chan types.Exec)
	currentStatus := types.StatusInitializing

	go func() {
		ticketCount := 0
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sessions, err := uwc.Exec.List(ctx, owner, projectName, listTerminated)
				if err != nil {
					var e *types.Error
					if errors.As(err, &e) {
						uie := &ui.Error{Error: e}
						fmt.Println(uie.Verbose())
						os.Exit(1)
					}
					errch <- err
					return
				}

				for _, s := range sessions {
					s := s
					if s.ID == execID {
						if s.Status != currentStatus {
							currentStatus = s.Status
							execch <- s
							return
						}
						if s.Status == types.StatusError {
							ui.Errorf("❌ Session %s failed to start", execID)
							os.Exit(1)
						}
						if s.Status == types.StatusTerminated {
							ui.Errorf("Session %q is terminated.", execID)
							os.Exit(1)
						}

						if ticketCount%10 == 0 && s.Status != types.StatusRunning {
							ui.Infof("Waiting for session %q to start...", execID)
						}
						ticketCount++
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return execch, errch, nil
}
