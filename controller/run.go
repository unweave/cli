package controller

import (
	"context"
	"fmt"
	"github.com/unweave/cli/entity"
	"io"
	"strings"
)

// Run runs the user's latest changes and environment with Unweave. It uploads the users
// code to the server and runs it. Any files/patterns in the .gitignore file will are
// from the upload.
func (c *Controller) Run(ctx context.Context) error {
	// get the root path of the user's currently active project (user must be inside subdirectory)
	// walk filesystem at root and zip every file that's not in .gitignore
	// create a new run-session - by making a call to api.unweave.io/compute/run-session
	// upload the zip file to the api.unweave.io/compute/run-session/upload/<rid> endpoint
	rid, err := c.api.CreateRunSession(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Created run session:", rid)

	// Walk the filesystem the repo root and zip up the files
	gatherFunc := gatherContext("~/<todo>/<path>")
	if err = c.api.UploadRunContext(ctx, rid, gatherFunc); err != nil {
		return err
	}

	return nil
}

// gatherContext zips up the user's code and environment and write it to a buffer to be
// uploaded to the server.
func gatherContext(rootDir string) entity.GatherContextFunc {
	// TODO: walk the filesystem and zip up the user's code
	return func(w io.Writer) error {
		reader := strings.NewReader("is anyone out there")
		_, err := io.Copy(w, reader)
		if err != nil {
			return err
		}
		return nil
	}

}
