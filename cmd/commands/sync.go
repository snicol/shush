package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Syncs all secrets from the storage provider to the cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(conf.Sync.Prefixes) == 0 {
				return fmt.Errorf("no prefixes listed for sync config on profile %s", profile)
			}

			return sess.Sync(context.Background(), conf.Sync.Prefixes)
		},
	}
)
