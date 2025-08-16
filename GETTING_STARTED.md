# Getting Started with Mirror CLI

## Quick Setup (5 minutes)

### 1. Build the CLI

```bash
cd mirror_cli
make build
```

### 2. Initialize Configuration

```bash
./build/mirror_cli config init
./build/mirror_cli config set --host localhost --port 8112
```

### 3. Test Connection

```bash
./build/mirror_cli peer list
```

## First Mirror Setup

### 1. Create Source Peer (PostgreSQL)

```bash
./build/mirror_cli peer create \
  --name my_postgres \
  --type postgres \
  --pg-host localhost \
  --pg-port 5432 \
  --pg-user postgres \
  --pg-password your_password \
  --pg-database your_db
```

### 2. Create Destination Peer (BigQuery)

```bash
./build/mirror_cli peer create \
  --name my_bigquery \
  --type bigquery \
  --bq-project your-project \
  --bq-dataset your_dataset \
  --bq-private-key "$(cat service-account.json | jq -r .private_key)" \
  --bq-client-email service@your-project.iam.gserviceaccount.com
```

### 3. Create Mirror

```bash
./build/mirror_cli mirror create \
  --name my_first_mirror \
  --source my_postgres \
  --destination my_bigquery \
  --tables "public.users->users,public.orders->orders" \
  --batch-size 1000
```

### 4. Monitor Mirror

```bash
./build/mirror_cli mirror status my_first_mirror
./build/mirror_cli mirror list
```

## Advanced Usage

### Automated Setup

Use the example scripts:

```bash
# Edit the script with your configuration
vim examples/setup-postgres-to-bigquery.sh

# Run the setup
./examples/setup-postgres-to-bigquery.sh
```

### Monitoring Dashboard

```bash
# Start the monitoring dashboard
./examples/monitor-mirrors.sh
```

### Mirror Management

```bash
# Pause a mirror
./build/mirror_cli mirror pause my_first_mirror

# Resume a mirror
./build/mirror_cli mirror resume my_first_mirror

# Add more tables
./build/mirror_cli mirror edit my_first_mirror \
  --add-tables "public.products->products"

# Drop a mirror
./build/mirror_cli mirror drop my_first_mirror --force
```

## Troubleshooting

### Connection Issues

1. **Check PeerDB is running**: `curl http://localhost:8113/v1/version`
2. **Verify gRPC port**: `telnet localhost 8112`
3. **Check configuration**: `./build/mirror_cli config show`

### Common Errors

- **"failed to connect"**: Verify host/port settings
- **"peer not found"**: Create peers before creating mirrors
- **"invalid table mapping"**: Use format `source_table->dest_table`

## Next Steps

1. **Install globally**: `make install`
2. **Set up monitoring**: Use the dashboard script
3. **Automate with scripts**: Customize the example scripts
4. **Add to CI/CD**: Use for automated deployments

## Help

- `./build/mirror_cli --help` - General help
- `./build/mirror_cli mirror --help` - Mirror commands
- `./build/mirror_cli peer --help` - Peer commands
- See [README.md](README.md) for complete documentation
