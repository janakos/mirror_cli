package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexcode/mirror_cli/internal/client"
	pb "github.com/alexcode/mirror_cli/proto/gen"
)

// peerCmd represents the peer command
var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Manage PeerDB peers",
	Long:  "Commands for creating, listing, and managing PeerDB peer connections.",
}

// peerListCmd represents the peer list command
var peerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all peers",
	Long:  "List all configured peer connections.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPeers(cmd)
	},
}

// peerCreateCmd represents the peer create command
var peerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new peer",
	Long:  "Create a new peer connection.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return createPeer(cmd)
	},
}

// peerDropCmd represents the peer drop command
var peerDropCmd = &cobra.Command{
	Use:   "drop [peer-name]",
	Short: "Drop a peer",
	Long:  "Drop a peer connection.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return dropPeer(cmd, args[0])
	},
}

// peerValidateCmd represents the peer validate command
var peerValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a peer configuration",
	Long:  "Validate a peer configuration without creating it.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return validatePeer(cmd)
	},
}

func init() {
	rootCmd.AddCommand(peerCmd)
	peerCmd.AddCommand(peerListCmd)
	peerCmd.AddCommand(peerCreateCmd)
	peerCmd.AddCommand(peerDropCmd)
	peerCmd.AddCommand(peerValidateCmd)

	// Create command flags
	addPeerCreateFlags(peerCreateCmd)
	addPeerCreateFlags(peerValidateCmd)

	// Create command specific flags
	peerCreateCmd.Flags().Bool("allow-update", false, "Allow updating existing peer")

	// Drop command flags
	peerDropCmd.Flags().Bool("force", false, "Force drop without confirmation")
}

func addPeerCreateFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Peer name (required)")
	cmd.Flags().String("type", "", "Peer type: postgres, bigquery, snowflake, etc. (required)")

	// PostgreSQL flags
	cmd.Flags().String("pg-host", "", "PostgreSQL host")
	cmd.Flags().Int("pg-port", 5432, "PostgreSQL port")
	cmd.Flags().String("pg-user", "", "PostgreSQL user")
	cmd.Flags().String("pg-password", "", "PostgreSQL password")
	cmd.Flags().String("pg-database", "", "PostgreSQL database")
	cmd.Flags().String("pg-tls-host", "", "PostgreSQL TLS host")
	cmd.Flags().String("pg-metadata-schema", "_peerdb_internal", "PostgreSQL metadata schema")

	// BigQuery flags
	cmd.Flags().String("bq-project", "", "BigQuery project ID")
	cmd.Flags().String("bq-dataset", "", "BigQuery dataset ID")
	cmd.Flags().String("bq-auth-type", "service_account", "BigQuery auth type")
	cmd.Flags().String("bq-private-key", "", "BigQuery private key")
	cmd.Flags().String("bq-private-key-id", "", "BigQuery private key ID")
	cmd.Flags().String("bq-client-email", "", "BigQuery client email")
	cmd.Flags().String("bq-client-id", "", "BigQuery client ID")

	// Snowflake flags
	cmd.Flags().String("sf-account", "", "Snowflake account ID")
	cmd.Flags().String("sf-user", "", "Snowflake username")
	cmd.Flags().String("sf-password", "", "Snowflake password")
	cmd.Flags().String("sf-private-key", "", "Snowflake private key")
	cmd.Flags().String("sf-database", "", "Snowflake database")
	cmd.Flags().String("sf-warehouse", "", "Snowflake warehouse")
	cmd.Flags().String("sf-role", "", "Snowflake role")
	cmd.Flags().String("sf-metadata-schema", "_PEERDB_INTERNAL", "Snowflake metadata schema")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")
}

func listPeers(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client
	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	// List peers
	resp, err := client.ListPeers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list peers: %w", err)
	}

	if len(resp.Items) == 0 {
		fmt.Println("No peers found")
		return nil
	}

	// Print header
	fmt.Printf("%-20s %-15s %-10s\n", "NAME", "TYPE", "CATEGORY")
	fmt.Println(strings.Repeat("-", 50))

	// Print all peers
	for _, peer := range resp.Items {
		category := "General"
		fmt.Printf("%-20s %-15s %-10s\n", peer.Name, peer.Type.String(), category)
	}

	// Print source peers if different
	if len(resp.SourceItems) > 0 && len(resp.SourceItems) != len(resp.Items) {
		fmt.Println("\nSource Peers:")
		for _, peer := range resp.SourceItems {
			fmt.Printf("%-20s %-15s %-10s\n", peer.Name, peer.Type.String(), "Source")
		}
	}

	// Print destination peers if different
	if len(resp.DestinationItems) > 0 && len(resp.DestinationItems) != len(resp.Items) {
		fmt.Println("\nDestination Peers:")
		for _, peer := range resp.DestinationItems {
			fmt.Printf("%-20s %-15s %-10s\n", peer.Name, peer.Type.String(), "Destination")
		}
	}

	return nil
}

func createPeer(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	name, _ := cmd.Flags().GetString("name")
	peerType, _ := cmd.Flags().GetString("type")
	allowUpdate, _ := cmd.Flags().GetBool("allow-update")

	// Create peer based on type
	peer, err := buildPeerFromFlags(cmd, name, peerType)
	if err != nil {
		return err
	}

	// Create client
	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	// Create the peer
	resp, err := client.CreatePeer(ctx, peer, allowUpdate)
	if err != nil {
		return fmt.Errorf("failed to create peer: %w", err)
	}

	status := "created"
	if resp.Status == pb.CreatePeerStatus_FAILED {
		status = "failed"
	}

	fmt.Printf("✓ Peer '%s' %s successfully\n", name, status)
	if resp.Message != "" {
		fmt.Printf("  Message: %s\n", resp.Message)
	}

	return nil
}

func validatePeer(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	name, _ := cmd.Flags().GetString("name")
	peerType, _ := cmd.Flags().GetString("type")

	// Create peer based on type
	peer, err := buildPeerFromFlags(cmd, name, peerType)
	if err != nil {
		return err
	}

	// Create client
	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	// Validate the peer
	resp, err := client.ValidatePeer(ctx, peer)
	if err != nil {
		return fmt.Errorf("failed to validate peer: %w", err)
	}

	status := "valid"
	if resp.Status == pb.ValidatePeerStatus_INVALID {
		status = "invalid"
	}

	fmt.Printf("Peer configuration is %s\n", status)
	if resp.Message != "" {
		fmt.Printf("Message: %s\n", resp.Message)
	}

	return nil
}

func dropPeer(cmd *cobra.Command, peerName string) error {
	force, _ := cmd.Flags().GetBool("force")

	// Confirmation unless forced
	if !force {
		fmt.Printf("Are you sure you want to drop peer '%s'? This action cannot be undone. (y/N): ", peerName)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.DropPeer(ctx, peerName); err != nil {
		return fmt.Errorf("failed to drop peer: %w", err)
	}

	fmt.Printf("✓ Peer '%s' dropped successfully\n", peerName)
	return nil
}

func buildPeerFromFlags(cmd *cobra.Command, name, peerType string) (*pb.Peer, error) {
	peer := &pb.Peer{
		Name: name,
	}

	switch strings.ToLower(peerType) {
	case "postgres", "postgresql":
		peer.Type = pb.DBType_POSTGRES
		config, err := buildPostgresConfig(cmd)
		if err != nil {
			return nil, err
		}
		peer.Config = &pb.Peer_PostgresConfig{PostgresConfig: config}

	case "bigquery", "bq":
		peer.Type = pb.DBType_BIGQUERY
		config, err := buildBigQueryConfig(cmd)
		if err != nil {
			return nil, err
		}
		peer.Config = &pb.Peer_BigqueryConfig{BigqueryConfig: config}

	case "snowflake", "sf":
		peer.Type = pb.DBType_SNOWFLAKE
		config, err := buildSnowflakeConfig(cmd)
		if err != nil {
			return nil, err
		}
		peer.Config = &pb.Peer_SnowflakeConfig{SnowflakeConfig: config}

	default:
		return nil, fmt.Errorf("unsupported peer type: %s", peerType)
	}

	return peer, nil
}

func buildPostgresConfig(cmd *cobra.Command) (*pb.PostgresConfig, error) {
	host, _ := cmd.Flags().GetString("pg-host")
	port, _ := cmd.Flags().GetInt("pg-port")
	user, _ := cmd.Flags().GetString("pg-user")
	password, _ := cmd.Flags().GetString("pg-password")
	database, _ := cmd.Flags().GetString("pg-database")
	tlsHost, _ := cmd.Flags().GetString("pg-tls-host")
	metadataSchema, _ := cmd.Flags().GetString("pg-metadata-schema")

	if host == "" || user == "" || database == "" {
		return nil, fmt.Errorf("postgres peer requires host, user, and database")
	}

	config := &pb.PostgresConfig{
		Host:     host,
		Port:     uint32(port),
		User:     user,
		Password: password,
		Database: database,
		TlsHost:  tlsHost,
	}

	if metadataSchema != "" {
		config.MetadataSchema = &metadataSchema
	}

	return config, nil
}

func buildBigQueryConfig(cmd *cobra.Command) (*pb.BigqueryConfig, error) {
	projectId, _ := cmd.Flags().GetString("bq-project")
	datasetId, _ := cmd.Flags().GetString("bq-dataset")
	authType, _ := cmd.Flags().GetString("bq-auth-type")
	privateKey, _ := cmd.Flags().GetString("bq-private-key")
	privateKeyId, _ := cmd.Flags().GetString("bq-private-key-id")
	clientEmail, _ := cmd.Flags().GetString("bq-client-email")
	clientId, _ := cmd.Flags().GetString("bq-client-id")

	if projectId == "" || datasetId == "" {
		return nil, fmt.Errorf("bigquery peer requires project and dataset")
	}

	return &pb.BigqueryConfig{
		AuthType:                authType,
		ProjectId:               projectId,
		PrivateKeyId:            privateKeyId,
		PrivateKey:              privateKey,
		ClientEmail:             clientEmail,
		ClientId:                clientId,
		AuthUri:                 "https://accounts.google.com/o/oauth2/auth",
		TokenUri:                "https://oauth2.googleapis.com/token",
		AuthProviderX509CertUrl: "https://www.googleapis.com/oauth2/v1/certs",
		DatasetId:               datasetId,
	}, nil
}

func buildSnowflakeConfig(cmd *cobra.Command) (*pb.SnowflakeConfig, error) {
	accountId, _ := cmd.Flags().GetString("sf-account")
	username, _ := cmd.Flags().GetString("sf-user")
	password, _ := cmd.Flags().GetString("sf-password")
	privateKey, _ := cmd.Flags().GetString("sf-private-key")
	database, _ := cmd.Flags().GetString("sf-database")
	warehouse, _ := cmd.Flags().GetString("sf-warehouse")
	role, _ := cmd.Flags().GetString("sf-role")
	metadataSchema, _ := cmd.Flags().GetString("sf-metadata-schema")

	if accountId == "" || username == "" || database == "" || warehouse == "" {
		return nil, fmt.Errorf("snowflake peer requires account, username, database, and warehouse")
	}

	if password == "" && privateKey == "" {
		return nil, fmt.Errorf("snowflake peer requires either password or private key")
	}

	config := &pb.SnowflakeConfig{
		AccountId:    accountId,
		Username:     username,
		Database:     database,
		Warehouse:    warehouse,
		Role:         role,
		QueryTimeout: 300, // 5 minutes default
	}

	if password != "" {
		config.Password = &password
	}
	if privateKey != "" {
		config.PrivateKey = privateKey
	}
	if metadataSchema != "" {
		config.MetadataSchema = &metadataSchema
	}

	return config, nil
}
