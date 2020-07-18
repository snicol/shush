package env

import (
	"errors"
	"net/url"
)

const shushSceheme = "shush"

// ParseURI validates env variable URIs and ensures the scheme matches
// It currently only extracts the desired 'key' segment. It could expand to use
// the 'port' field for versioning
func ParseURI(rawurl string) (string, error) {
	url, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return "", err
	}

	if url.Scheme != shushSceheme {
		return "", errors.New("invalid scheme for URI")
	}

	return url.Hostname(), nil
}
