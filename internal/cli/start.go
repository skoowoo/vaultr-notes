package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hardhacker/vaultr/internal/agent"
	"github.com/hardhacker/vaultr/internal/config"
	"github.com/hardhacker/vaultr/internal/logger"
	"github.com/hardhacker/vaultr/internal/server"
	"github.com/hardhacker/vaultr/internal/storage"
	"github.com/spf13/cobra"
)

func defaultServerPIDFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".vaultr", "pid")
	}
	return filepath.Join(home, ".vaultr", "pid")
}

// set by `vaultr start server --pid-file`
var servePIDFile string

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start a service",
		SilenceUsage: true,
	}

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start the database server",
		Long: `Start the database server.

The listening address (TCP host/port, timeouts) is read only from
the vaultr.toml config file — not from flags or environment variables.

By default the server listens on TCP at 127.0.0.1:54321. Set server.port to 0
in vaultr.toml to disable the server entirely.

After the server successfully binds its listen address, its process ID is written
to the pid file (see --pid-file). The file is removed when the server exits.`,
		SilenceUsage: true,
		RunE:         runServe,
	}
	serverCmd.Flags().StringVar(&servePIDFile, "pid-file", defaultServerPIDFile(),
		"path to write this process ID after the server starts listening (use empty string to skip)")

	cmd.AddCommand(serverCmd)

	return cmd
}

func runServe(_ *cobra.Command, _ []string) error {
	agent.WarmShellEnv() // async: captures login-shell env for agent spawns

	cfg, cfgFileUsed, err := config.Load("")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfgFileUsed != "" {
		if abs, aerr := filepath.Abs(cfgFileUsed); aerr == nil {
			cfgFileUsed = abs
		}
	}

	log := logger.New(cfg.Log)

	vault, err := storage.New(cfg.Vault.Path)
	if err != nil {
		return fmt.Errorf("open vault: %w", err)
	}
	log.Info("vault opened", "path", vault.Root())

	pidFile := strings.TrimSpace(servePIDFile)

	srv := server.New(cfg, cfgFileUsed, log, vault)
	return srv.Run(pidFile)
}
