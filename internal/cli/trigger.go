package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "trigger",
		Short:        "Manually trigger server-side operations",
		SilenceUsage: true,
	}

	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Trigger an immediate git sync",
		Long: `Trigger an immediate pull-then-push cycle on the running server.

Requires the git_sync plugin to be enabled (plugins.git_sync.enabled = true
in vaultr.toml) and the server to be running.

Example:
  vaultr trigger git`,
		SilenceUsage: true,
		RunE:         runGitSync,
	}

	cmd.AddCommand(gitCmd)

	return cmd
}

func runGitSync(_ *cobra.Command, _ []string) error {
	c, err := openClient()
	if err != nil {
		return err
	}
	if err := c.TriggerGitSync(); err != nil {
		return fmt.Errorf("git sync: %w", err)
	}
	fmt.Println("sync requested")
	return nil
}
