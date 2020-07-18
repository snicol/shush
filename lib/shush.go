package shush

import (
	"context"
	"errors"

	"shush/lib/cache"
	"shush/lib/storage"
)

// UpsertVersionBehaviour allows control over what happens when a newer version
// in the storage provider is found.
type UpsertVersionBehaviour int

const (
	// UpsertVersionReplaceDifferent if the version is different, it will be replaced
	UpsertVersionReplaceDifferent UpsertVersionBehaviour = iota
	// UpsertVersionReplaceNewer only if the latest version is newer will it replace
	UpsertVersionReplaceNewer
	// UpsertVersionSkipCheck ensures that checks for the latest version on the
	// storage provider skipped entirely. Useful for users relying on the
	// sync command
	UpsertVersionSkipCheck
)

type Session struct {
	storage storage.Provider
	cache   cache.Provider

	upsertBehaviour UpsertVersionBehaviour
}

func NewSession(storage storage.Provider, cache cache.Provider, upsertBehaviour UpsertVersionBehaviour) *Session {
	return &Session{
		storage:         storage,
		cache:           cache,
		upsertBehaviour: upsertBehaviour,
	}
}

func (s *Session) Get(ctx context.Context, key string) (val string, ver int, err error) {
	val, ver, found, err := s.getCache(ctx, key)
	if err != nil {
		return "", 0, err
	}

	if found {
		return val, ver, nil
	}

	vs, err := s.storage.Get(ctx, []string{key})
	if err != nil {
		return "", 0, err
	}

	if len(vs) != 1 {
		return "", 0, errors.New("unexpected number of results")
	}

	liveVal, liveVer := vs[0].Value, vs[0].Version

	if err := s.setCache(ctx, liveVer, key, liveVal); err != nil {
		return "", 0, err
	}

	return liveVal, liveVer, nil
}

func (s *Session) getCache(ctx context.Context, key string) (string, int, bool, error) {
	if s.cache == nil {
		return "", 0, false, nil
	}

	cacheVal, cacheVer, err := s.cache.Get(key)
	if err != nil && err != cache.ErrNotFound {
		return "", 0, false, err
	}

	if s.upsertBehaviour == UpsertVersionSkipCheck {
		return cacheVal, cacheVer, true, nil
	}

	latestLiveVersion, err := s.storage.LatestVersion(ctx, key)
	if err != nil {
		return "", 0, false, err
	}

	if latestLiveVersion == cacheVer {
		return cacheVal, cacheVer, true, nil
	}

	if s.upsertBehaviour == UpsertVersionReplaceNewer && latestLiveVersion <= cacheVer {
		return cacheVal, cacheVer, true, nil
	}

	return "", 0, false, nil
}

func (s *Session) Set(ctx context.Context, key, value string) error {
	return s.storage.Set(ctx, key, value)
}

func (s *Session) setCache(ctx context.Context, version int, k, v string) error {
	if s.cache == nil {
		return nil
	}

	return s.cache.Set(version, k, v)
}

func (s *Session) Sync(ctx context.Context, prefixes []string) error {
	syncStorage, ok := s.storage.(storage.SyncableProvider)
	if !ok {
		return errors.New("sync is not implemented on this storage type")
	}

	if s.cache == nil {
		return errors.New("sync called with no cache provider")
	}

	for _, prefix := range prefixes {
		keys, err := syncStorage.GetByPrefix(ctx, prefix)
		if err != nil {
			return err
		}

		if len(keys) == 0 {
			continue
		}

		// TODO(sn): look into making caches match the same interface as
		// storage for Get, so that session.Get and this loop can be
		// simplified
		for _, key := range keys {
			_, _, err = s.Get(ctx, key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
