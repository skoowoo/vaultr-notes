package cli

import (
	"fmt"
	"os"

	"github.com/hardhacker/vaultr/internal/build"
	"github.com/hardhacker/vaultr/internal/client"
	"github.com/hardhacker/vaultr/internal/config"
	"github.com/hardhacker/vaultr/internal/storage"
	"github.com/spf13/cobra"
)

// rootCmd is the base command. Without subcommands it prints command help.
var rootCmd = &cobra.Command{
	Use:           "vaultr",
	Short:         "AI-native personal note-taking system",
	Long:          `AI-native personal note-taking system`,
	SilenceUsage:  true,
	SilenceErrors: true, // we print errors ourselves in Execute for full control
	Args:          cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := cmd.Flags().GetBool("version")
		if ver {
			fmt.Println(build.Get())
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "print version information")

	// Register subcommands.
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newStartCmd())
	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newStatusCmd())
	rootCmd.AddCommand(newAppendCmd())
	rootCmd.AddCommand(newPrependCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newReadCmd())
	rootCmd.AddCommand(newSearchCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newKnowledgeCmd())
	rootCmd.AddCommand(newResolveCmd())
	rootCmd.AddCommand(newExtractCmd())
	rootCmd.AddCommand(newTagCmd())
	rootCmd.AddCommand(newTriggerCmd())
	rootCmd.AddCommand(newAgentCmd())
	rootCmd.AddCommand(newShortCmd())
}

// Execute is the single entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// openClient loads config and returns an HTTP client connected to the running
// Vaultr server. All data-access CLI commands use this instead of openVault.
func openClient() (*client.Client, error) {
	cfg := config.MustLoad("")
	return client.New(cfg)
}

// openVault loads config and opens the Vault directly.
// Only used by the "serve" command itself, which needs direct filesystem access.
func openVault() (*storage.Vault, error) {
	cfg := config.MustLoad("")
	return storage.New(cfg.Vault.Path)
}
