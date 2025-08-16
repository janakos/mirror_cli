package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/alexcode/mirror_cli/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mirror_cli",
	Short: "A CLI for managing PeerDB mirrors via gRPC",
	Long: `mirror_cli is a command-line interface for managing PeerDB mirrors using gRPC.
	
It provides commands to create, list, pause, resume, drop, and monitor mirrors,
as well as manage peer connections.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(loadConfigFile)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mirror_cli/config.yaml)")
	rootCmd.PersistentFlags().String("host", "localhost", "PeerDB server host")
	rootCmd.PersistentFlags().Int("port", 8112, "PeerDB server port")
	rootCmd.PersistentFlags().Bool("tls", false, "Use TLS connection")
	rootCmd.PersistentFlags().String("username", "", "Username for authentication")
	rootCmd.PersistentFlags().String("password", "", "Password for authentication")

	// Bind flags to viper
	viper.BindPFlag("peerdb_host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("peerdb_port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("tls", rootCmd.PersistentFlags().Lookup("tls"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
}

// loadConfigFile reads in config file and ENV variables if set.
func loadConfigFile() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".mirror_cli" (without extension).
		viper.AddConfigPath(home + "/.mirror_cli")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// GetConfig returns the loaded configuration
func GetConfig() *config.Config {
	return cfg
}
