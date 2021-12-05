package controller

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/unweave/cli/entity"
	"gopkg.in/gookit/color.v1"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const defaultGitIgnore = `
.git
`

// Run runs the user's latest changes and environment with Unweave. It uploads the users
// code to the server and runs it. Any files/patterns in the .gitignore file will are
// from the upload.
func (c *Controller) Run(ctx context.Context, path string) error {
	if err := c.cfg.ValidateProjectDir(path); err != nil {
		msg := "Ow snap! Looks like you don't have a currently active Unweave project. \n" +
			"Either switch to a unweave project folder or create a new one by running: \n" +
			color.Blue.Render("unweave init")
		fmt.Println(msg)
		return err
	}

	rid, err := c.api.CreateRunSession(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Created run session:", rid)

	// Walk the filesystem the repo root and zip up the files
	gatherFunc := gatherContext(path)
	if err = c.api.UploadRunContext(ctx, rid, gatherFunc); err != nil {
		return err
	}

	return nil
}

// gatherContext zips up the user's code and environment and write it to a buffer to be
// uploaded to the server.
func gatherContext(rootDir string) entity.GatherContextFunc {
	giPath := filepath.Join(rootDir, ".gitignore")
	lines := strings.Split(defaultGitIgnore, "\n")

	// Compile ignore pattern - use GitIgnore if it exists
	var gi *ignore.GitIgnore
	if _, err := os.Stat(giPath); os.IsNotExist(err) {
		gi = ignore.CompileIgnoreLines(lines...)
	} else {
		gi, err = ignore.CompileIgnoreFileAndLines(giPath, lines...)
		if err != nil {
			fmt.Println("Error compiling .gitignore file:", err)
			fmt.Println("Ignoring .gitignore file")
		}
	}

	return func(w io.Writer) error {
		zw := gzip.NewWriter(w)
		tw := tar.NewWriter(zw)

		err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if gi.MatchesPath(path) {
				return nil
			}
			if d.IsDir() {
				return nil
			}

			f, err := d.Info()
			if err != nil {
				return err
			}

			// generate tar header
			header, err := tar.FileInfoHeader(f, path)
			if err != nil {
				return err
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			header.Name = filepath.ToSlash(path)
			if err = tw.WriteHeader(header); err != nil {
				return err
			}

			data, err := os.Open(path)
			if err != nil {
				return err
			}
			if _, err = io.Copy(tw, data); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return err
		}
		if err = tw.Close(); err != nil {
			return err
		}
		if err = zw.Close(); err != nil {
			return err
		}
		return nil
	}
}
