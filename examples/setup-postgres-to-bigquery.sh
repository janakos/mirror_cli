#!/bin/bash

# Example script: Set up a complete PostgreSQL to BigQuery mirror
# This demonstrates the full workflow from peer creation to mirror setup

set -e

echo "🚀 Setting up PostgreSQL to BigQuery mirror..."

# Configuration
POSTGRES_PEER="postgres_source"
BIGQUERY_PEER="bigquery_dest"
MIRROR_NAME="postgres_to_bq_mirror"

# PostgreSQL connection details
POSTGRES_HOST="localhost"
POSTGRES_PORT=5432
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="password"
POSTGRES_DB="sourcedb"

# BigQuery connection details
BQ_PROJECT="my-project-id"
BQ_DATASET="replicated_data"
BQ_SERVICE_ACCOUNT_KEY="./service-account.json"

echo "📋 Step 1: Initialize CLI configuration"
mirror_cli config init --force
mirror_cli config set --host localhost --port 8112

echo "📋 Step 2: Create PostgreSQL source peer"
mirror_cli peer create \
  --name "$POSTGRES_PEER" \
  --type postgres \
  --pg-host "$POSTGRES_HOST" \
  --pg-port "$POSTGRES_PORT" \
  --pg-user "$POSTGRES_USER" \
  --pg-password "$POSTGRES_PASSWORD" \
  --pg-database "$POSTGRES_DB"

echo "📋 Step 3: Create BigQuery destination peer"
if [ ! -f "$BQ_SERVICE_ACCOUNT_KEY" ]; then
  echo "❌ Error: BigQuery service account key file not found at $BQ_SERVICE_ACCOUNT_KEY"
  echo "Please download your service account key and update the BQ_SERVICE_ACCOUNT_KEY variable"
  exit 1
fi

# Extract key details from service account JSON
BQ_PRIVATE_KEY=$(cat "$BQ_SERVICE_ACCOUNT_KEY" | jq -r '.private_key')
BQ_CLIENT_EMAIL=$(cat "$BQ_SERVICE_ACCOUNT_KEY" | jq -r '.client_email')
BQ_PRIVATE_KEY_ID=$(cat "$BQ_SERVICE_ACCOUNT_KEY" | jq -r '.private_key_id')
BQ_CLIENT_ID=$(cat "$BQ_SERVICE_ACCOUNT_KEY" | jq -r '.client_id')

mirror_cli peer create \
  --name "$BIGQUERY_PEER" \
  --type bigquery \
  --bq-project "$BQ_PROJECT" \
  --bq-dataset "$BQ_DATASET" \
  --bq-private-key "$BQ_PRIVATE_KEY" \
  --bq-client-email "$BQ_CLIENT_EMAIL" \
  --bq-private-key-id "$BQ_PRIVATE_KEY_ID" \
  --bq-client-id "$BQ_CLIENT_ID"

echo "📋 Step 4: Validate peer connections"
echo "Testing PostgreSQL connection..."
mirror_cli peer validate \
  --name "test_postgres" \
  --type postgres \
  --pg-host "$POSTGRES_HOST" \
  --pg-port "$POSTGRES_PORT" \
  --pg-user "$POSTGRES_USER" \
  --pg-password "$POSTGRES_PASSWORD" \
  --pg-database "$POSTGRES_DB"

echo "Testing BigQuery connection..."
mirror_cli peer validate \
  --name "test_bigquery" \
  --type bigquery \
  --bq-project "$BQ_PROJECT" \
  --bq-dataset "$BQ_DATASET" \
  --bq-private-key "$BQ_PRIVATE_KEY" \
  --bq-client-email "$BQ_CLIENT_EMAIL"

echo "📋 Step 5: Create CDC mirror"
mirror_cli mirror create \
  --name "$MIRROR_NAME" \
  --source "$POSTGRES_PEER" \
  --destination "$BIGQUERY_PEER" \
  --tables "public.users->users,public.orders->orders,public.products->products" \
  --batch-size 1000 \
  --idle-timeout 60 \
  --publication "peerdb_pub" \
  --replication-slot "peerdb_slot"

echo "📋 Step 6: Monitor mirror status"
echo "Waiting for mirror to initialize..."
sleep 5

mirror_cli mirror status "$MIRROR_NAME"

echo "✅ Mirror setup complete!"
echo ""
echo "📊 Next steps:"
echo "1. Monitor your mirror: mirror_cli mirror status $MIRROR_NAME"
echo "2. List all mirrors: mirror_cli mirror list"
echo "3. Pause if needed: mirror_cli mirror pause $MIRROR_NAME"
echo "4. Resume: mirror_cli mirror resume $MIRROR_NAME"
echo ""
echo "🔄 Your data is now replicating from PostgreSQL to BigQuery!"
