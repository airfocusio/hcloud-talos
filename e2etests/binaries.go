package e2etests

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

type BinariesUnpack interface {
	Done() bool
	Unpack(reader io.Reader) error
}

var _ BinariesUnpack = (*RawBinariesUnpack)(nil)

type RawBinariesUnpack struct {
	Name string
}

func (u *RawBinariesUnpack) Done() bool {
	if stat, err := os.Stat(u.Name); err != nil || stat.IsDir() {
		return false
	}
	return true
}

func (u *RawBinariesUnpack) Unpack(reader io.Reader) error {
	if err := os.MkdirAll(path.Dir(u.Name), 0o755); err != nil {
		return err
	}
	fmt.Printf("Writing binary to %s\n", u.Name)
	file, err := os.OpenFile(u.Name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}
	return nil
}

var _ BinariesUnpack = (*TarGzBinariesUnpack)(nil)

type TarGzBinariesUnpack struct {
	Entries map[string]string
}

func (u *TarGzBinariesUnpack) Done() bool {
	for _, name := range u.Entries {
		if stat, err := os.Stat(name); err != nil || stat.IsDir() {
			return false
		}
	}
	return true
}

func (u *TarGzBinariesUnpack) Unpack(reader io.Reader) error {
	gzipArchive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzipArchive.Close()

	tarArchive := tar.NewReader(gzipArchive)
	for {
		header, err := tarArchive.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if name, ok := u.Entries[header.Name]; ok {
			if err := os.MkdirAll(path.Dir(name), 0o755); err != nil {
				return err
			}
			fmt.Printf("Writing binary to %s\n", name)
			file, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(file, tarArchive)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func PrepareBinaries(url string, unpack BinariesUnpack) error {
	if unpack.Done() {
		return nil
	}

	fmt.Printf("Downloading %s\n", url)
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := unpack.Unpack(resp.Body); err != nil {
		return err
	}

	return nil
}
