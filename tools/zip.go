package tools

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
)

func Gzip(rootDir string, w io.Writer, ignore *ignore.GitIgnore) error {
	zw := gzip.NewWriter(w)
	tw := tar.NewWriter(zw)

	defer zw.Close()
	defer tw.Close()

	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		if ignore.MatchesPath(rPath) {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}

		// generate tar header
		header, err := tar.FileInfoHeader(fi, rPath)
		if err != nil {
			return err
		}

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(rPath)
		if err = tw.WriteHeader(header); err != nil {
			return err
		}

		data, err := os.Open(path)
		if err != nil {
			return err
		}
		defer data.Close()

		if _, err = io.Copy(tw, data); err != nil {
			return err
		}
		return nil
	})
}

func Zip(rootDir string, w io.Writer, ignore *ignore.GitIgnore) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		if ignore.MatchesPath(rPath) {
			return nil
		}
		if d.IsDir() {
			rPath += "/" // Add a trailing slash for directories
		}

		fw, err := zw.Create(rPath)
		if err != nil {
			return err
		}

		if !d.IsDir() {
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(fw, data)
			data.Close()
			if err != nil {
				return err
			}
		}
		return nil
	})
}
