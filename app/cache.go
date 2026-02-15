package app

import (
	"encoding/base64"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type cache struct {
	a fyne.App

	enc base64.Encoding
}

func makeCache(a fyne.App) fyne.Cache {
	c := &cache{a: a, enc: base64.Encoding{}}

	root := c.RootURI()
	exists, err := storage.Exists(root)
	if !exists || err != nil {
		err = storage.CreateListable(root)
		if err != nil {
			fyne.LogError("Failed to create app cache space", err)
		}
	}

	return c
}

func (c *cache) RootURI() fyne.URI {
	return storage.NewFileURI(rootCacheDir(c.a))
}

func (c *cache) Exists(name string) bool {
	path := c.encodePath(name)

	ok, err := storage.Exists(path)
	return ok && err == nil
}

func (c *cache) Get(name string) ([]byte, error) {
	path := c.encodePath(name)

	read, err := storage.Reader(path)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(read)
}

func (c *cache) Set(name string, data []byte) error {
	path := c.encodePath(name)

	write, err := storage.Writer(path)
	if err != nil {
		return err
	}

	n, err := write.Write(data)
	for n < len(data) {
		if err != nil {
			return err
		}

		data = data[n:]
		n, err = write.Write(data)
	}

	return err
}

func (c *cache) Remove(name string) error {
	path := c.encodePath(name)

	return storage.Delete(path)
}

func (c *cache) encodePath(badName string) fyne.URI {
	name := base64.StdEncoding.EncodeToString([]byte(badName))
	child, _ := storage.Child(c.a.Cache().RootURI(), name)
	return child
}
