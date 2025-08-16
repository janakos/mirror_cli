# Configuration Files

This directory contains YAML configuration files for PeerDB mirrors and peers. These files can be version controlled and applied to manage your infrastructure as code.

## Directory Structure

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

## Usage

### Export Existing Configurations

```bash
# Export a peer configuration
mirror_cli config export-peer my_postgres --output configs/peers/production/postgres.yaml

# Export a mirror configuration
mirror_cli config export-mirror my_mirror --output configs/mirrors/production/users-sync.yaml
```

### Apply Configurations

```bash
# Apply a peer configuration
mirror_cli config apply -f configs/peers/production/postgres.yaml

# Apply a mirror configuration
mirror_cli config apply -f configs/mirrors/production/users-sync.yaml

# Apply all configurations in a directory
mirror_cli config apply -f configs/peers/production/
```

### Validate Configurations

```bash
# Validate a configuration file
mirror_cli config validate -f configs/peers/production/postgres.yaml
```

## Configuration File Format

See the `examples/` directory for sample configuration files.
