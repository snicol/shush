package commands

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

var (
	setCmd = &cobra.Command{
		Use:   "set",
		Short: "Writes a secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("not enough args")
			}

			return sess.Set(context.Background(), args[0], args[1])
		},
	}
)
