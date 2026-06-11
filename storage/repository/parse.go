package repository

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
)

var (
	rxHost = regexp.MustCompile(`^(.*?)(?::[0-9]+)?$`)
	rxName = regexp.MustCompile(`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`)
)

// NewFileURI implements the back-end logic to storage.NewFileURI, which you
// should use instead. This is only here because other functions in repository
// need to call it, and it prevents a circular import.
//
// Since: 2.0
func NewFileURI(path string) fyne.URI {
	// URIs are supposed to use forward slashes. On Windows, it
	// should be OK to use the platform native filepath with UNIX
	// or NT style paths, with / or \, but when we reconstruct
	// the URI, we want to have / only.
	if runtime.GOOS == "windows" {
		// seems that sometimes we end up with
		// double-backslashes
		path = filepath.ToSlash(path)
	}

	return &uri{url.URL{
		Scheme: "file",
		Path:   path,
	}}
}

// ParseURI implements the back-end logic for storage.ParseURI, which you
// should use instead. This is only here because other functions in repository
// need to call it, and it prevents a circular import.
//
// Since: 2.0
func ParseURI(s string) (fyne.URI, error) {
	// Extract the scheme.
	scheme, path, ok := strings.Cut(s, ":")
	if !ok {
		return nil, errors.New("invalid URI, scheme must be present")
	}

	if strings.EqualFold(scheme, "file") {
		// Does this really deserve to be special? In principle, the
		// purpose of this check is to pass it to NewFileURI, which
		// allows platform path seps in the URI (against the RFC, but
		// easier for people building URIs naively on Windows). Maybe
		// we should punt this to whoever generated the URI in the
		// first place?

		if len(path) <= 2 { // I.e. file: and // given we know scheme.
			return nil, errors.New("not a valid URI")
		}

		if path[:2] == "//" {
			path = path[2:]
		}

		p, err := url.PathUnescape(path)
		if err != nil {
			return nil, err
		}

		// Windows files can break authority checks, so just return the parsed file URI
		return NewFileURI(p), nil
	}

	if strings.EqualFold(scheme, "urn") {
		return &uri{url.URL{
			Scheme: scheme,
			Path:   path,
		}}, nil
	}

	scheme = strings.ToLower(scheme)
	repo, err := ForScheme(scheme)
	if err == nil {
		// If the repository registered for this scheme implements a parser
		if c, ok := repo.(CustomURIRepository); ok {
			return c.ParseURI(s)
		}
	}

	// There was no repository registered, or it did not provide a parser

	l, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	if l.Host == "" {
		return &uri{*l}, nil
	}

	// host format checks
	m := rxHost.FindStringSubmatch(l.Host)
	if len(m) != 2 {
		return nil, errors.New("failed to find host")
	}

	if rxName.MatchString(m[1]) {
		return &uri{*l}, nil
	}

	if len(m[1]) > 1 && m[1][0] == '[' && m[1][len(m[1])-1] == ']' {
		m[1] = m[1][1 : len(m[1])-1]
	}

	if net.ParseIP(m[1]) == nil {
		return nil, fmt.Errorf("invalid address: %v", m[1])
	}

	return &uri{*l}, nil
}
