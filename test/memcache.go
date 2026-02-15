package test

import (
	"errors"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type memCache struct {
	memStore map[string][]byte
}

func makeCache() fyne.Cache {
	return &memCache{memStore: make(map[string][]byte)}
}

func (c *memCache) RootURI() fyne.URI {
	return storage.NewFileURI(os.TempDir()) // in case anyone wants to manually handle storage
}

func (c *memCache) Exists(name string) bool {
	_, ok := c.memStore[name]
	return ok
}

func (c *memCache) Get(name string) ([]byte, error) {
	data, ok := c.memStore[name]
	if !ok {
		return nil, errors.New("not found")
	}

	return data, nil
}

func (c *memCache) Set(name string, data []byte) error {
	c.memStore[name] = data
	return nil
}

func (c *memCache) Remove(name string) error {
	delete(c.memStore, name)
	return nil
}
