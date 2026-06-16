package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	var table bool

	cmd := &cobra.Command{
		Use:          "info",
		Short:        "Show server configuration and plugin status",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := openClient()
			if err != nil {
				return err
			}
			info, err := c.Info()
			if err != nil {
				return err
			}
			if table {
				fmt.Println(info.FormatTable())
				return nil
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(info)
		},
	}

	cmd.Flags().BoolVarP(&table, "table", "t", false, "output in human-readable format")

	return cmd
}
