#!/bin/bash

# Example workflow for setting up peers and mirrors from configuration files
# This demonstrates how to use the configuration management features

set -e

echo "🚀 Setting up PeerDB from configuration files..."

# Set environment variables (replace with your actual values)
export POSTGRES_PASSWORD="your_postgres_password"
export SNOWFLAKE_ACCOUNT="your_account.region"
export SNOWFLAKE_PRIVATE_KEY="your_rsa_private_key"

echo ""
echo "📋 Step 1: Validate all configurations"
./build/mirror_cli config validate -f configs/peers/development/
./build/mirror_cli config validate -f configs/mirrors/development/

echo ""
echo "🔧 Step 2: Apply peer configurations"
./build/mirror_cli config apply -f configs/peers/development/ --dry-run
read -p "Apply peer configurations? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    ./build/mirror_cli config apply -f configs/peers/development/
    echo "✅ Peers created successfully"
else
    echo "❌ Skipping peer creation"
    exit 0
fi

echo ""
echo "📊 Step 3: Verify peers are working"
./build/mirror_cli peer list

echo ""
echo "🔄 Step 4: Apply mirror configurations"
./build/mirror_cli config apply -f configs/mirrors/development/ --dry-run
read -p "Apply mirror configurations? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    ./build/mirror_cli config apply -f configs/mirrors/development/
    echo "✅ Mirrors created successfully"
else
    echo "❌ Skipping mirror creation"
    exit 0
fi

echo ""
echo "📈 Step 5: Check mirror status"
./build/mirror_cli mirror list
./build/mirror_cli mirror status dev_test_sync

echo ""
echo "🎉 Setup complete! Your mirrors are now configured and running."
echo ""
echo "💡 Tips:"
echo "   - Edit configs in configs/ directory and apply changes with 'config apply'"
echo "   - Use 'config export-peer' and 'config export-mirror' to save existing configs"
echo "   - Check mirror status with 'mirror status <name>'"
echo "   - Monitor logs and metrics through PeerDB UI"
