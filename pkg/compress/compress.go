package compress

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	ignore "github.com/sabhiram/go-gitignore"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
		if ignore.MatchesPath(path) {
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
		rPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
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
		if ignore.MatchesPath(path) {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		data, err := os.Open(path)
		if err != nil {
			return err
		}
		defer data.Close()

		rPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		fw, err := zw.Create(rPath)
		if err != nil {
			return err
		}

		if _, err = io.Copy(fw, data); err != nil {
			return err
		}
		return nil
	})
}
