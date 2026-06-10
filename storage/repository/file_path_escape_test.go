package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURI_filePathEscape(t *testing.T) {
	assert.Equal(t, "/home/user/file.txt", filePathEscape("/home/user/file.txt"))
	assert.Equal(t, "/home/user/file%231.txt", filePathEscape("/home/user/file#1.txt"))
}
