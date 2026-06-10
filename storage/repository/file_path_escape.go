package repository

//go:generate go run dont_escape_gen.go

const upperhex = "0123456789ABCDEF"

func filePathEscape(path string) string {
	length := len(path)
	for _, c := range []byte(path) {
		if !dontEscape[c] {
			length += 2
		}
	}
	if length == len(path) {
		return path
	}

	r := make([]byte, length)
	n := 0
	for _, c := range []byte(path) {
		if dontEscape[c] {
			r[n] = c
			n++
		} else {
			r[n] = '%'
			r[n+1] = upperhex[c>>4]
			r[n+2] = upperhex[c&15]
			n += 3
		}
	}
	return string(r)
}
