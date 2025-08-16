# Mirror CLI

A powerful command-line interface for managing PeerDB mirrors via gRPC. This CLI provides complete control over PeerDB mirror operations including creation, monitoring, editing, and deletion of mirrors and peer connections.

## Features

- **Mirror Management**: Create, list, pause, resume, drop, and edit CDC mirrors
- **Peer Management**: Create, list, validate, and drop peer connections
- **Configuration Management**: Easy configuration with YAML files and environment variables
- **Multiple Database Support**: PostgreSQL, BigQuery, Snowflake, and more
- **Real-time Status**: Monitor mirror status and progress
- **Cross-platform**: Available for Linux, macOS, and Windows

## Quick Start

### Prerequisites

- Go 1.21 or later
- [buf](https://buf.build/docs/installation) CLI tool for protobuf generation
- Access to a running PeerDB instance

### Installation

#### Option 1: Build from Source

```bash
# Clone and build
git clone <your-repo>
cd mirror_cli
make build

# Install to $GOPATH/bin
make install
```

#### Option 2: Download Release Binary

```bash
# Download the latest release for your platform
curl -L https://github.com/your-org/mirror_cli/releases/latest/download/mirror_cli-$(uname -s)-$(uname -m) -o mirror_cli
chmod +x mirror_cli
sudo mv mirror_cli /usr/local/bin/
```

### Initial Setup

1. **Initialize configuration**:
```bash
mirror_cli config init
```

2. **Configure PeerDB connection**:
```bash
mirror_cli config set --host localhost --port 8112
```

3. **Test connection**:
```bash
mirror_cli peer list
```

## Infrastructure as Code with Configuration Files

**Mirror CLI** supports managing peers and mirrors through YAML configuration files, enabling GitOps workflows and version control of your data replication infrastructure.

### Configuration Structure

```
configs/
├── peers/
│   ├── production/
│   ├── staging/
│   └── development/
├── mirrors/
│   ├── production/
│   ├── staging/
│   └── development/
└── examples/
    ├── peer-examples/
    └── mirror-examples/
```

### Configuration File Examples

**PostgreSQL Peer Configuration:**
```yaml
apiVersion: v1
kind: Peer
metadata:
  name: postgres_source
  environment: production
  description: Primary PostgreSQL database
spec:
  type: postgres
  config:
    host: postgres.company.com
    port: 5432
    user: peerdb_user
    password: ${POSTGRES_PASSWORD}
    database: users_db
    metadata_schema: _peerdb_internal
```

**Snowflake Peer Configuration:**
```yaml
apiVersion: v1
kind: Peer
metadata:
  name: snowflake_warehouse
  environment: production
  description: Snowflake data warehouse
spec:
  type: snowflake
  config:
    account_id: ${SNOWFLAKE_ACCOUNT}
    username: peerdb_user
    private_key: ${SNOWFLAKE_PRIVATE_KEY}
    database: ANALYTICS_DB
    warehouse: COMPUTE_WH
    role: PEERDB_ROLE
```

**CDC Mirror Configuration:**
```yaml
apiVersion: v1
kind: Mirror
metadata:
  name: users_sync_mirror
  environment: production
  description: Sync user data from PostgreSQL to Snowflake
spec:
  type: cdc
  source: postgres_source
  destination: snowflake_warehouse
  tables:
    - source: public.users
      destination: ANALYTICS_DB.PUBLIC.USERS
      partition_key: created_at
    - source: public.user_profiles
      destination: ANALYTICS_DB.PUBLIC.USER_PROFILES
      exclude_columns:
        - password_hash
        - ssn
  cdc:
    batch_size: 1000
    idle_timeout_seconds: 60
    initial_snapshot: true
    publication_name: peerdb_users_pub
    replication_slot_name: peerdb_users_slot
```

### Configuration Management Commands

```bash
# Validate configuration files
mirror_cli config validate -f configs/peers/production/
mirror_cli config validate -f configs/mirrors/production/users-sync.yaml

# Apply configurations (with dry-run first)
mirror_cli config apply -f configs/peers/production/ --dry-run
mirror_cli config apply -f configs/peers/production/

# Apply single configuration
mirror_cli config apply -f configs/mirrors/production/users-sync.yaml

# Export existing configurations
mirror_cli config export-peer my_postgres --output configs/peers/production/postgres.yaml
mirror_cli config export-mirror my_mirror --output configs/mirrors/production/users-sync.yaml
```

### GitOps Workflow

1. **Define Infrastructure**: Create YAML configurations in `configs/`
2. **Version Control**: Commit configurations to git
3. **Validate**: Run `config validate` in CI/CD pipelines
4. **Apply**: Use `config apply` to deploy changes
5. **Monitor**: Check status with `mirror status`

### Environment Variables

Configuration files support environment variable substitution using `${VAR_NAME}` syntax:

```bash
export POSTGRES_PASSWORD="secure_password"
export SNOWFLAKE_ACCOUNT="myaccount.us-east-1"
export SNOWFLAKE_PRIVATE_KEY="$(cat private_key.pem)"
```

## CLI Configuration

The CLI uses a YAML configuration file located at `~/.mirror_cli/config.yaml`. You can also use environment variables or command-line flags.

### Configuration Methods (in order of precedence):

1. **Command-line flags**: `--host`, `--port`, `--tls`
2. **Environment variables**: `MIRROR_CLI_PEERDB_HOST`, `MIRROR_CLI_PEERDB_PORT`, `MIRROR_CLI_TLS`
3. **Configuration file**: `~/.mirror_cli/config.yaml`

### Example Configuration File

```yaml
peerdb_host: "localhost"
peerdb_port: 8112
tls: false
username: ""
password: ""
```

## Usage Examples

### Peer Management

#### Create a PostgreSQL Peer

```bash
mirror_cli peer create \
  --name my_postgres \
  --type postgres \
  --pg-host localhost \
  --pg-port 5432 \
  --pg-user postgres \
  --pg-password mypassword \
  --pg-database mydb
```

#### Create a BigQuery Peer

```bash
mirror_cli peer create \
  --name my_bigquery \
  --type bigquery \
  --bq-project my-project \
  --bq-dataset my_dataset \
  --bq-private-key "$(cat service-account.json | jq -r .private_key)" \
  --bq-client-email service@my-project.iam.gserviceaccount.com
```

#### Create a Snowflake Peer

```bash
mirror_cli peer create \
  --name my_snowflake \
  --type snowflake \
  --sf-account myaccount.us-east-1 \
  --sf-user myuser \
  --sf-password mypassword \
  --sf-database MYDB \
  --sf-warehouse COMPUTE_WH \
  --sf-role ACCOUNTADMIN
```

#### List Peers

```bash
mirror_cli peer list
```

#### Validate Peer Configuration

```bash
mirror_cli peer validate \
  --name test_peer \
  --type postgres \
  --pg-host localhost \
  --pg-user postgres \
  --pg-database testdb
```

#### Drop a Peer

```bash
mirror_cli peer drop my_postgres --force
```

### Mirror Management

#### Create a CDC Mirror

```bash
mirror_cli mirror create \
  --name my_cdc_mirror \
  --source my_postgres \
  --destination my_bigquery \
  --tables "public.users->dataset.users,public.orders->dataset.orders" \
  --batch-size 1000 \
  --idle-timeout 60 \
  --publication peerdb_pub \
  --replication-slot peerdb_slot
```

#### List Mirrors

```bash
mirror_cli mirror list
```

#### Get Mirror Status

```bash
mirror_cli mirror status my_cdc_mirror
```

#### Pause a Mirror

```bash
mirror_cli mirror pause my_cdc_mirror
```

#### Resume a Mirror

```bash
mirror_cli mirror resume my_cdc_mirror
```

#### Edit Mirror Configuration

```bash
# Add new tables
mirror_cli mirror edit my_cdc_mirror \
  --add-tables "public.products->dataset.products" \
  --batch-size 2000

# Remove tables
mirror_cli mirror edit my_cdc_mirror \
  --remove-tables "public.old_table->dataset.old_table"
```

#### Drop a Mirror

```bash
mirror_cli mirror drop my_cdc_mirror --force
```

### Configuration Commands

#### Show Current Configuration

```bash
mirror_cli config show
```

#### Update Configuration

```bash
mirror_cli config set --host production.peerdb.com --port 8112 --tls
```

#### Initialize New Configuration

```bash
mirror_cli config init --force
```

## Command Reference

### Global Flags

- `--config`: Config file path (default: `~/.mirror_cli/config.yaml`)
- `--host`: PeerDB server host (default: `localhost`)
- `--port`: PeerDB server port (default: `8112`)
- `--tls`: Use TLS connection
- `--username`: Username for authentication
- `--password`: Password for authentication

### Mirror Commands

| Command | Description |
|---------|-------------|
| `mirror create` | Create a new CDC mirror |
| `mirror list` | List all mirrors |
| `mirror status` | Get detailed mirror status |
| `mirror pause` | Pause a running mirror |
| `mirror resume` | Resume a paused mirror |
| `mirror edit` | Edit mirror configuration |
| `mirror drop` | Drop a mirror permanently |

### Peer Commands

| Command | Description |
|---------|-------------|
| `peer create` | Create a new peer connection |
| `peer list` | List all peer connections |
| `peer validate` | Validate peer configuration |
| `peer drop` | Drop a peer connection |

### Config Commands

| Command | Description |
|---------|-------------|
| `config show` | Show current CLI configuration |
| `config set` | Set CLI configuration values |
| `config init` | Initialize new CLI configuration |
| `config apply` | Apply peer/mirror configurations from files |
| `config validate` | Validate configuration files |
| `config export-peer` | Export peer configuration to file |
| `config export-mirror` | Export mirror configuration to file |

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Generate protobuf files
make proto

# Install dependencies
make deps
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Lint code
make lint
```

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit your changes: `git commit -am 'Add feature'`
7. Push to the branch: `git push origin feature-name`
8. Submit a pull request

## Troubleshooting

### Common Issues

1. **Connection Failed**
   ```
   Error: failed to connect to PeerDB at localhost:8112
   ```
   - Verify PeerDB is running on the specified host and port
   - Check if the gRPC port (8112) is accessible
   - Ensure firewall rules allow the connection

2. **Authentication Failed**
   - Verify username and password are correct
   - Check if authentication is required for your PeerDB instance

3. **Invalid Peer Configuration**
   - Use `peer validate` command to check configuration before creating
   - Verify connection details (host, port, credentials)
   - Check database permissions

4. **Mirror Creation Failed**
   - Ensure source and destination peers exist and are valid
   - Verify table names and mapping format
   - Check if replication slot and publication exist (for PostgreSQL)

### Getting Help

- Use `mirror_cli --help` for general help
- Use `mirror_cli [command] --help` for command-specific help
- Check the [PeerDB documentation](https://docs.peerdb.io/) for more details

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built for [PeerDB](https://peerdb.io/) - Real-time data movement platform
- Uses [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [Viper](https://github.com/spf13/viper) for configuration management
