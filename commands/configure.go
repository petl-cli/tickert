package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure authentication and default settings",
	Long: `Stores credentials and preferences in ~/.config/discovery-api/config.yaml.

Environment variables always take precedence over this file.
Supported keys: bearer_token, api_key, base_url, output_format`,
	RunE: runConfigure,
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

func runConfigure(cmd *cobra.Command, args []string) error {
	path, err := _configLoader.ConfigFilePath()
	if err != nil {
		return err
	}
	fmt.Printf("Config file: %s\n\n", path)
	fmt.Println("Set credentials via environment variables or edit the config file directly.")
	fmt.Println()
	fmt.Printf("  API key:      export %s=<your-key>\n", "DISCOVERY_API_API_KEY")
	fmt.Printf("  Base URL:     export %s_BASE_URL=<url>\n", "DISCOVERY_API")
	return nil
}
