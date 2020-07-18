package shush

import (
	"context"
	"os"

	"shush/lib/env"
)

// GetenvContext gets the shush URI value for a given ENV variable key and
// validates/extracts the secret key from it. It then returns the decrypted
// value.
func (s *Session) GetenvContext(ctx context.Context, key string) (string, error) {
	envVal := os.Getenv(key)

	parsedKey, err := env.ParseURI(envVal)
	if err != nil {
		return "", err
	}

	value, _, err := s.Get(ctx, parsedKey)
	return value, err
}

// Getenv calls GetenvContext with a default background context. Use
// GetenvContext where possible.
func (s *Session) Getenv(key string) (string, error) {
	ctx := context.Background()

	return s.GetenvContext(ctx, key)
}
