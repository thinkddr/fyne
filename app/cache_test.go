package app

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"github.com/stretchr/testify/assert"
)

func TestCache_Exists(t *testing.T) {
	c := NewWithID("io.fyne.test").Cache()
	testName := "entry"

	assert.False(t, c.Exists(testName))

	err := c.Set(testName, []byte("data"))
	assert.NoError(t, err)

	assert.True(t, c.Exists(testName))

	// and verify it was stored
	c = NewWithID("io.fyne.test").Cache()
	assert.True(t, c.Exists(testName))

	_ = c.Remove(testName)
}

func TestCache_ExistsTest(t *testing.T) {
	c := test.NewApp().Cache()
	testName := "entry"

	assert.False(t, c.Exists(testName))

	err := c.Set(testName, []byte("data"))
	assert.NoError(t, err)

	assert.True(t, c.Exists(testName))
	// no need to remove a test cache, it resets
}

func TestCache_Set(t *testing.T) {
	c := NewWithID("io.fyne.test").Cache()
	testName := "entry2"

	err := c.Set(testName, []byte("data"))
	assert.NoError(t, err)

	data, err := c.Get(testName)
	assert.NoError(t, err)
	assert.Equal(t, []byte("data"), data)

	// and verify it was stored
	c = NewWithID("io.fyne.test").Cache()

	data, err = c.Get(testName)
	assert.NoError(t, err)
	assert.Equal(t, []byte("data"), data)

	_ = c.Remove(testName)
}

func TestCache_SetTest(t *testing.T) {
	c := test.NewApp().Cache()
	testName := "entry2"

	err := c.Set(testName, []byte("data"))
	assert.NoError(t, err)

	data, err := c.Get(testName)
	assert.NoError(t, err)
	assert.Equal(t, []byte("data"), data)
}
