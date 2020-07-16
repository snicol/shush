package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Gets a secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("not enough args, expected key")
			}

			value, ver, err := sess.Get(context.Background(), args[0])
			if err != nil {
				return err
			}

			fmt.Printf("version: %d\nvalue: %s\n", ver, value)
			return nil
		},
	}
)
