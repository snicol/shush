package cache

import (
	"errors"
)

var ErrNotFound = errors.New("not found")

type Provider interface {
	Get(k string) (string, int, error)
	Set(version int, k, v string) error
}
