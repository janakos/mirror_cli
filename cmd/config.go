package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/janakos/mirror_cli/internal/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  "Commands for managing the CLI configuration settings.",
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current CLI configuration settings.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return showConfig()
	},
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Long:  "Set configuration values and save them to the config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return setConfig(cmd)
	},
}

// configInitCmd represents the config init command
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  "Initialize a new configuration file with default values.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configInitCmd)

	// Set command flags
	configSetCmd.Flags().String("host", "", "PeerDB server host")
	configSetCmd.Flags().Int("port", 0, "PeerDB server port")
	configSetCmd.Flags().Bool("tls", false, "Use TLS connection")
	configSetCmd.Flags().String("username", "", "Username for authentication")
	configSetCmd.Flags().String("password", "", "Password for authentication")

	// Init command flags
	configInitCmd.Flags().Bool("force", false, "Overwrite existing config file")
}

func showConfig() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Println("Current Configuration:")
	fmt.Printf("  Host:     %s\n", cfg.PeerDBHost)
	fmt.Printf("  Port:     %d\n", cfg.PeerDBPort)
	fmt.Printf("  TLS:      %t\n", cfg.TLS)
	fmt.Printf("  Username: %s\n", cfg.Username)
	fmt.Printf("  Address:  %s\n", cfg.Address())

	if cfg.Password != "" {
		fmt.Printf("  Password: [set]\n")
	} else {
		fmt.Printf("  Password: [not set]\n")
	}

	return nil
}

func setConfig(cmd *cobra.Command) error {
	// Load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Update values from flags
	if cmd.Flags().Changed("host") {
		host, _ := cmd.Flags().GetString("host")
		cfg.PeerDBHost = host
		fmt.Printf("Set host to: %s\n", host)
	}

	if cmd.Flags().Changed("port") {
		port, _ := cmd.Flags().GetInt("port")
		cfg.PeerDBPort = port
		fmt.Printf("Set port to: %d\n", port)
	}

	if cmd.Flags().Changed("tls") {
		tls, _ := cmd.Flags().GetBool("tls")
		cfg.TLS = tls
		fmt.Printf("Set TLS to: %t\n", tls)
	}

	if cmd.Flags().Changed("username") {
		username, _ := cmd.Flags().GetString("username")
		cfg.Username = username
		fmt.Printf("Set username to: %s\n", username)
	}

	if cmd.Flags().Changed("password") {
		password, _ := cmd.Flags().GetString("password")
		cfg.Password = password
		fmt.Println("Set password: [hidden]")
	}

	// Save the configuration
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✓ Configuration saved successfully")
	return nil
}

func initializeConfig(cmd *cobra.Command) error {
	force, _ := cmd.Flags().GetBool("force")

	// Check if config already exists
	if !force {
		if _, err := config.LoadConfig(); err == nil {
			fmt.Println("Configuration file already exists. Use --force to overwrite.")
			return nil
		}
	}

	// Create default config
	cfg := config.DefaultConfig()

	// Save the configuration
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	fmt.Println("✓ Configuration initialized with default values")
	fmt.Printf("  Config saved to: ~/.mirror_cli/config.yaml\n")
	fmt.Printf("  Default host: %s\n", cfg.PeerDBHost)
	fmt.Printf("  Default port: %d\n", cfg.PeerDBPort)
	fmt.Printf("\nYou can modify these settings using 'mirror_cli config set' or by editing the config file directly.\n")

	return nil
}
