package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/janakos/mirror_cli/internal/client"
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

// configApplyCmd represents the config apply command
var configApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply configuration from file(s)",
	Long:  "Apply peer and mirror configurations from YAML files.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return applyConfigs(cmd)
	},
}

// configValidateCmd represents the config validate command
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file(s)",
	Long:  "Validate peer and mirror configuration files without applying them.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return validateConfigs(cmd)
	},
}

// configExportPeerCmd represents the config export-peer command
var configExportPeerCmd = &cobra.Command{
	Use:   "export-peer [peer-name]",
	Short: "Export peer configuration to file",
	Long:  "Export an existing peer configuration to a YAML file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exportPeerConfig(cmd, args[0])
	},
}

// configExportMirrorCmd represents the config export-mirror command
var configExportMirrorCmd = &cobra.Command{
	Use:   "export-mirror [mirror-name]",
	Short: "Export mirror configuration to file",
	Long:  "Export an existing mirror configuration to a YAML file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exportMirrorConfig(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configApplyCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configExportPeerCmd)
	configCmd.AddCommand(configExportMirrorCmd)

	// Set command flags
	configSetCmd.Flags().String("host", "", "PeerDB server host")
	configSetCmd.Flags().Int("port", 0, "PeerDB server port")
	configSetCmd.Flags().Bool("tls", false, "Use TLS connection")
	configSetCmd.Flags().String("username", "", "Username for authentication")
	configSetCmd.Flags().String("password", "", "Password for authentication")

	// Init command flags
	configInitCmd.Flags().Bool("force", false, "Overwrite existing config file")

	// Apply command flags
	configApplyCmd.Flags().StringP("file", "f", "", "Configuration file or directory path")
	configApplyCmd.Flags().Bool("dry-run", false, "Show what would be applied without actually applying")
	configApplyCmd.Flags().Bool("force", false, "Force apply even if resources already exist")
	configApplyCmd.MarkFlagRequired("file")

	// Validate command flags
	configValidateCmd.Flags().StringP("file", "f", "", "Configuration file or directory path")
	configValidateCmd.MarkFlagRequired("file")

	// Export peer command flags
	configExportPeerCmd.Flags().StringP("output", "o", "", "Output file path")
	configExportPeerCmd.Flags().String("environment", "production", "Environment to set in metadata")

	// Export mirror command flags
	configExportMirrorCmd.Flags().StringP("output", "o", "", "Output file path")
	configExportMirrorCmd.Flags().String("environment", "production", "Environment to set in metadata")
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

func applyConfigs(cmd *cobra.Command) error {
	filePath, _ := cmd.Flags().GetString("file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Check if path is a file or directory
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access path %s: %w", filePath, err)
	}

	var configs []*config.FileConfig
	if fileInfo.IsDir() {
		configs, err = config.LoadConfigsFromDirectory(filePath)
		if err != nil {
			return fmt.Errorf("failed to load configs from directory: %w", err)
		}
	} else {
		cfg, err := config.LoadConfigFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
		configs = []*config.FileConfig{cfg}
	}

	if len(configs) == 0 {
		fmt.Println("No configuration files found")
		return nil
	}

	// Create client for applying configurations
	var grpcClient *client.Client
	if !dryRun {
		grpcClient, err = client.NewClient(GetConfig())
		if err != nil {
			return fmt.Errorf("failed to create gRPC client: %w", err)
		}
		defer grpcClient.Close()
	}

	// Apply each configuration
	for _, cfg := range configs {
		fmt.Printf("Processing %s '%s'...\n", cfg.Kind, cfg.Metadata.Name)

		if dryRun {
			fmt.Printf("  [DRY-RUN] Would apply %s configuration\n", cfg.Kind)
			continue
		}

		switch cfg.Kind {
		case "Peer":
			err = applyPeerConfig(ctx, grpcClient, cfg, force)
		case "Mirror":
			err = applyMirrorConfig(ctx, grpcClient, cfg, force)
		default:
			err = fmt.Errorf("unsupported configuration kind: %s", cfg.Kind)
		}

		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
			return err
		}
		fmt.Printf("  ✅ Applied successfully\n")
	}

	if dryRun {
		fmt.Printf("\n[DRY-RUN] %d configurations would be applied\n", len(configs))
	} else {
		fmt.Printf("\n✅ Successfully applied %d configurations\n", len(configs))
	}

	return nil
}

func validateConfigs(cmd *cobra.Command) error {
	filePath, _ := cmd.Flags().GetString("file")

	// Check if path is a file or directory
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access path %s: %w", filePath, err)
	}

	var configs []*config.FileConfig
	if fileInfo.IsDir() {
		configs, err = config.LoadConfigsFromDirectory(filePath)
		if err != nil {
			return fmt.Errorf("failed to load configs from directory: %w", err)
		}
	} else {
		cfg, err := config.LoadConfigFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
		configs = []*config.FileConfig{cfg}
	}

	if len(configs) == 0 {
		fmt.Println("No configuration files found")
		return nil
	}

	allValid := true
	for _, cfg := range configs {
		fmt.Printf("Validating %s '%s'...\n", cfg.Kind, cfg.Metadata.Name)

		var err error
		switch cfg.Kind {
		case "Peer":
			_, err = cfg.ToPeerProto()
		case "Mirror":
			_, err = cfg.ToMirrorProto()
		default:
			err = fmt.Errorf("unsupported configuration kind: %s", cfg.Kind)
		}

		if err != nil {
			fmt.Printf("  ❌ Invalid: %v\n", err)
			allValid = false
		} else {
			fmt.Printf("  ✅ Valid\n")
		}
	}

	if allValid {
		fmt.Printf("\n✅ All %d configurations are valid\n", len(configs))
	} else {
		fmt.Printf("\n❌ Some configurations are invalid\n")
		return fmt.Errorf("validation failed")
	}

	return nil
}

func exportPeerConfig(cmd *cobra.Command, peerName string) error {
	output, _ := cmd.Flags().GetString("output")
	environment, _ := cmd.Flags().GetString("environment")

	// Default output path if not specified
	if output == "" {
		output = fmt.Sprintf("configs/peers/%s/%s.yaml", environment, peerName)
	}

	fmt.Printf("Exporting peer '%s' to %s...\n", peerName, output)
	
	// TODO: Implement peer export by fetching from PeerDB
	// For now, create a template
	fileConfig := &config.FileConfig{
		APIVersion: "v1",
		Kind:       "Peer",
		Metadata: config.Metadata{
			Name:        peerName,
			Environment: environment,
			Description: fmt.Sprintf("Configuration for %s peer", peerName),
		},
		Spec: config.Spec{
			Type: "postgres", // Default, should be detected
			Config: config.PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "${POSTGRES_PASSWORD}",
				Database: "mydb",
			},
		},
	}

	if err := config.SaveConfigFile(fileConfig, output); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Peer configuration exported to %s\n", output)
	fmt.Printf("💡 Note: Update the configuration with actual values before applying\n")

	return nil
}

func exportMirrorConfig(cmd *cobra.Command, mirrorName string) error {
	output, _ := cmd.Flags().GetString("output")
	environment, _ := cmd.Flags().GetString("environment")

	// Default output path if not specified
	if output == "" {
		output = fmt.Sprintf("configs/mirrors/%s/%s.yaml", environment, mirrorName)
	}

	fmt.Printf("Exporting mirror '%s' to %s...\n", mirrorName, output)
	
	// TODO: Implement mirror export by fetching from PeerDB
	// For now, create a template
	fileConfig := &config.FileConfig{
		APIVersion: "v1",
		Kind:       "Mirror",
		Metadata: config.Metadata{
			Name:        mirrorName,
			Environment: environment,
			Description: fmt.Sprintf("Configuration for %s mirror", mirrorName),
		},
		Spec: config.Spec{
			Type:        "cdc",
			Source:      "postgres_source",
			Destination: "snowflake_warehouse",
			Tables: []config.TableConfig{
				{
					Source:      "public.example_table",
					Destination: "ANALYTICS_DB.PUBLIC.EXAMPLE_TABLE",
				},
			},
			CDC: &config.CDCConfig{
				BatchSize:           1000,
				IdleTimeoutSeconds:  60,
				InitialSnapshot:     true,
				PublicationName:     "peerdb_pub",
				ReplicationSlotName: "peerdb_slot",
			},
		},
	}

	if err := config.SaveConfigFile(fileConfig, output); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✅ Mirror configuration exported to %s\n", output)
	fmt.Printf("💡 Note: Update the configuration with actual values before applying\n")

	return nil
}

func applyPeerConfig(ctx context.Context, grpcClient *client.Client, cfg *config.FileConfig, force bool) error {
	peer, err := cfg.ToPeerProto()
	if err != nil {
		return fmt.Errorf("failed to convert config to peer: %w", err)
	}

	_, err = grpcClient.CreatePeer(ctx, peer, force)
	return err
}

func applyMirrorConfig(ctx context.Context, grpcClient *client.Client, cfg *config.FileConfig, force bool) error {
	mirrorReq, err := cfg.ToMirrorProto()
	if err != nil {
		return fmt.Errorf("failed to convert config to mirror: %w", err)
	}

	_, err = grpcClient.CreateCDCMirror(ctx, mirrorReq)
	return err
}
