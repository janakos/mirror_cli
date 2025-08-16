package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	pb "github.com/janakos/mirror_cli/proto/gen"
)

// FileConfig represents a configuration file structure
type FileConfig struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

// Metadata contains configuration metadata
type Metadata struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Spec contains the configuration specification
type Spec struct {
	// For Peer configurations
	Type       string      `yaml:"type,omitempty"`
	Config     interface{} `yaml:"config,omitempty"`
	Validation *Validation `yaml:"validation,omitempty"`

	// For Mirror configurations
	Source      string        `yaml:"source,omitempty"`
	Destination string        `yaml:"destination,omitempty"`
	Tables      []TableConfig `yaml:"tables,omitempty"`
	CDC         *CDCConfig    `yaml:"cdc,omitempty"`
	Snapshot    *SnapshotConfig `yaml:"snapshot,omitempty"`
	Columns     *ColumnsConfig  `yaml:"columns,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
}

// Validation contains validation settings
type Validation struct {
	Timeout       string `yaml:"timeout,omitempty"`
	RetryAttempts int    `yaml:"retry_attempts,omitempty"`
}

// TableConfig represents table mapping configuration
type TableConfig struct {
	Source           string   `yaml:"source"`
	Destination      string   `yaml:"destination"`
	PartitionKey     string   `yaml:"partition_key,omitempty"`
	ExcludeColumns   []string `yaml:"exclude_columns,omitempty"`
}

// CDCConfig contains CDC-specific configuration
type CDCConfig struct {
	BatchSize             uint32 `yaml:"batch_size,omitempty"`
	IdleTimeoutSeconds    uint64 `yaml:"idle_timeout_seconds,omitempty"`
	InitialSnapshot       bool   `yaml:"initial_snapshot,omitempty"`
	PublicationName       string `yaml:"publication_name,omitempty"`
	ReplicationSlotName   string `yaml:"replication_slot_name,omitempty"`
}

// SnapshotConfig contains snapshot-specific configuration
type SnapshotConfig struct {
	NumRowsPerPartition    uint32 `yaml:"num_rows_per_partition,omitempty"`
	MaxParallelWorkers     uint32 `yaml:"max_parallel_workers,omitempty"`
	NumTablesInParallel    uint32 `yaml:"num_tables_in_parallel,omitempty"`
}

// ColumnsConfig contains column-specific configuration
type ColumnsConfig struct {
	SoftDeleteColumn string `yaml:"soft_delete_column,omitempty"`
	SyncedAtColumn   string `yaml:"synced_at_column,omitempty"`
}

// PostgresConfig represents PostgreSQL configuration
type PostgresConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Database       string `yaml:"database"`
	TLSHost        string `yaml:"tls_host,omitempty"`
	MetadataSchema string `yaml:"metadata_schema,omitempty"`
}

// SnowflakeConfig represents Snowflake configuration
type SnowflakeConfig struct {
	AccountID      string `yaml:"account_id"`
	Username       string `yaml:"username"`
	PrivateKey     string `yaml:"private_key,omitempty"`
	Password       string `yaml:"password,omitempty"`
	Database       string `yaml:"database"`
	Warehouse      string `yaml:"warehouse"`
	Role           string `yaml:"role,omitempty"`
	QueryTimeout   uint64 `yaml:"query_timeout,omitempty"`
	MetadataSchema string `yaml:"metadata_schema,omitempty"`
}

// LoadConfigFile loads a configuration file from disk
func LoadConfigFile(filename string) (*FileConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	content := os.ExpandEnv(string(data))

	var config FileConfig
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// SaveConfigFile saves a configuration to disk
func SaveConfigFile(config *FileConfig, filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ToPeerProto converts a FileConfig to a Peer protobuf
func (fc *FileConfig) ToPeerProto() (*pb.Peer, error) {
	if fc.Kind != "Peer" {
		return nil, fmt.Errorf("config is not a Peer, got: %s", fc.Kind)
	}

	peer := &pb.Peer{
		Name: fc.Metadata.Name,
	}

	switch strings.ToLower(fc.Spec.Type) {
	case "postgres", "postgresql":
		peer.Type = pb.DBType_POSTGRES
		pgConfig, err := convertToPostgresConfig(fc.Spec.Config)
		if err != nil {
			return nil, err
		}
		peer.Config = &pb.Peer_PostgresConfig{PostgresConfig: pgConfig}

	case "snowflake":
		peer.Type = pb.DBType_SNOWFLAKE
		sfConfig, err := convertToSnowflakeConfig(fc.Spec.Config)
		if err != nil {
			return nil, err
		}
		peer.Config = &pb.Peer_SnowflakeConfig{SnowflakeConfig: sfConfig}

	default:
		return nil, fmt.Errorf("unsupported peer type: %s", fc.Spec.Type)
	}

	return peer, nil
}

// ToMirrorProto converts a FileConfig to mirror creation request
func (fc *FileConfig) ToMirrorProto() (*pb.CreateCDCFlowRequest, error) {
	if fc.Kind != "Mirror" {
		return nil, fmt.Errorf("config is not a Mirror, got: %s", fc.Kind)
	}

	// Convert table mappings
	tableMappings := make([]*pb.TableMapping, len(fc.Spec.Tables))
	for i, table := range fc.Spec.Tables {
		tableMappings[i] = &pb.TableMapping{
			SourceTableIdentifier:      table.Source,
			DestinationTableIdentifier: table.Destination,
			PartitionKey:               table.PartitionKey,
			Exclude:                    table.ExcludeColumns,
		}
	}

	// Build connection config
	connectionConfig := &pb.FlowConnectionConfigs{
		FlowJobName:         fc.Metadata.Name,
		SourceName:          fc.Spec.Source,
		DestinationName:     fc.Spec.Destination,
		TableMappings:       tableMappings,
		Env:                 fc.Spec.Env,
	}

	// Add CDC configuration
	if fc.Spec.CDC != nil {
		connectionConfig.MaxBatchSize = fc.Spec.CDC.BatchSize
		connectionConfig.IdleTimeoutSeconds = fc.Spec.CDC.IdleTimeoutSeconds
		connectionConfig.DoInitialSnapshot = fc.Spec.CDC.InitialSnapshot
		connectionConfig.PublicationName = fc.Spec.CDC.PublicationName
		connectionConfig.ReplicationSlotName = fc.Spec.CDC.ReplicationSlotName
	}

	// Add snapshot configuration
	if fc.Spec.Snapshot != nil {
		connectionConfig.SnapshotNumRowsPerPartition = fc.Spec.Snapshot.NumRowsPerPartition
		connectionConfig.SnapshotMaxParallelWorkers = fc.Spec.Snapshot.MaxParallelWorkers
		connectionConfig.SnapshotNumTablesInParallel = fc.Spec.Snapshot.NumTablesInParallel
	}

	// Add column configuration
	if fc.Spec.Columns != nil {
		connectionConfig.SoftDeleteColName = fc.Spec.Columns.SoftDeleteColumn
		connectionConfig.SyncedAtColName = fc.Spec.Columns.SyncedAtColumn
	}

	return &pb.CreateCDCFlowRequest{
		ConnectionConfigs: connectionConfig,
	}, nil
}

// convertToPostgresConfig converts interface{} to PostgresConfig
func convertToPostgresConfig(config interface{}) (*pb.PostgresConfig, error) {
	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	var pgConfig PostgresConfig
	if err := yaml.Unmarshal(data, &pgConfig); err != nil {
		return nil, err
	}

	pbConfig := &pb.PostgresConfig{
		Host:     pgConfig.Host,
		Port:     uint32(pgConfig.Port),
		User:     pgConfig.User,
		Password: pgConfig.Password,
		Database: pgConfig.Database,
		TlsHost:  pgConfig.TLSHost,
	}

	if pgConfig.MetadataSchema != "" {
		pbConfig.MetadataSchema = &pgConfig.MetadataSchema
	}

	return pbConfig, nil
}

// convertToSnowflakeConfig converts interface{} to SnowflakeConfig
func convertToSnowflakeConfig(config interface{}) (*pb.SnowflakeConfig, error) {
	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	var sfConfig SnowflakeConfig
	if err := yaml.Unmarshal(data, &sfConfig); err != nil {
		return nil, err
	}

	pbConfig := &pb.SnowflakeConfig{
		AccountId:    sfConfig.AccountID,
		Username:     sfConfig.Username,
		Database:     sfConfig.Database,
		Warehouse:    sfConfig.Warehouse,
		Role:         sfConfig.Role,
		QueryTimeout: sfConfig.QueryTimeout,
	}

	if sfConfig.PrivateKey != "" {
		pbConfig.PrivateKey = sfConfig.PrivateKey
	}
	if sfConfig.Password != "" {
		pbConfig.Password = &sfConfig.Password
	}
	if sfConfig.MetadataSchema != "" {
		pbConfig.MetadataSchema = &sfConfig.MetadataSchema
	}

	return pbConfig, nil
}

// LoadConfigsFromDirectory loads all config files from a directory
func LoadConfigsFromDirectory(dirPath string) ([]*FileConfig, error) {
	var configs []*FileConfig

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml") {
			config, err := LoadConfigFile(path)
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", path, err)
			}
			configs = append(configs, config)
		}

		return nil
	})

	return configs, err
}
