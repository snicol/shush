package storage

import (
	"context"
)

type Provider interface {
	Get(ctx context.Context, keys []string) ([]Result, error)
	Set(ctx context.Context, key, value string) error
	LatestVersion(ctx context.Context, key string) (int, error)
}

type Result struct {
	Value   string
	Version int
}
