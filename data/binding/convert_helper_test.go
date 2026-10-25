package binding

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2/storage"
)

func TestParseIntBounds(t *testing.T) {
	value, err := parseInt("0x10")
	assert.NoError(t, err)
	assert.Equal(t, 16, value)

	tooLarge := "2147483648"
	if strconv.IntSize == 64 {
		tooLarge = "9223372036854775808"
	}
	_, err = parseInt(tooLarge)
	assert.Error(t, err)
}

func TestStripPrecision(t *testing.T) {
	format := "%2.3f"
	assert.Equal(t, "%2f", stripFormatPrecision(format))
	format = "total=%3.4f%%"
	assert.Equal(t, "total=%3f%%", stripFormatPrecision(format))

	format = "%.2f"
	assert.Equal(t, "%f", stripFormatPrecision(format))

	format = "%4d"
	assert.Equal(t, "%4d", stripFormatPrecision(format))

	format = "%v"
	assert.Equal(t, "%v", stripFormatPrecision(format))

	format = "%.f"
	assert.Equal(t, "%f", stripFormatPrecision(format))

	format = "foo=%.f"
	assert.Equal(t, "foo=%f", stripFormatPrecision(format))
}

func TestURIFromStringHelper(t *testing.T) {
	str := "file:///tmp/test.txt"
	u, err := uriFromString(str)

	assert.Nil(t, err)
	assert.Equal(t, str, u.String())
}

func TestURIToStringHelper(t *testing.T) {
	u := storage.NewFileURI("/tmp/test.txt")
	str, err := uriToString(u)

	assert.Nil(t, err)
	assert.Equal(t, u.String(), str)

	str, err = uriToString(nil)
	assert.Nil(t, err)
	assert.Equal(t, "", str)
}
