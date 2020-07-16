package shush

import (
	"context"
	"errors"

	"shush/lib/cache"
	"shush/lib/storage"
)

// UpsertVersionBehaviour allows control over what happens when a newer version
// in the storage provider is
type UpsertVersionBehaviour int

const (
	// UpsertVersionReplaceDifferent if the version is different, it will be replaced
	UpsertVersionReplaceDifferent UpsertVersionBehaviour = 0
	// UpsertVersionReplaceNewer only if the latest version is newer will it replace
	UpsertVersionReplaceNewer = 1
)

type Session struct {
	cache   cache.Provider
	storage storage.Provider

	upsertBehaviour UpsertVersionBehaviour
}

func NewSession(cache cache.Provider, storage storage.Provider, upsertBehaviour UpsertVersionBehaviour) *Session {
	return &Session{
		cache:           cache,
		storage:         storage,
		upsertBehaviour: upsertBehaviour,
	}
}

func (s *Session) Get(ctx context.Context, key string) (val string, ver int, err error) {
	cacheVal, cacheVer, err := s.cache.Get(key)
	if err != nil && err != cache.ErrNotFound {
		return "", 0, err
	}

	latestLiveVersion, err := s.storage.LatestVersion(ctx, key)
	if err != nil {
		return "", 0, err
	}

	if latestLiveVersion == cacheVer {
		return cacheVal, cacheVer, nil
	}

	if s.upsertBehaviour == UpsertVersionReplaceNewer && latestLiveVersion <= cacheVer {
		return cacheVal, cacheVer, nil
	}

	vs, err := s.storage.Get(ctx, []string{key})
	if err != nil {
		return "", 0, err
	}

	if len(vs) != 1 {
		return "", 0, errors.New("unexpected number of results")
	}

	liveVal, liveVer := vs[0].Value, vs[0].Version

	err = s.cache.Set(liveVer, key, liveVal)
	if err != nil {
		return "", 0, err
	}

	return liveVal, liveVer, nil
}

func (s *Session) Set(ctx context.Context, key, value string) error {
	return s.storage.Set(ctx, key, value)
}
