#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 Railzway Auth - Database Setup (Remote Nomad/Docker)${NC}"
echo "========================================================"

# Configuration
read -p "Enter server IP [34.87.70.45]: " SERVER_IP
SERVER_IP=${SERVER_IP:-34.87.70.45}
read -p "Enter SSH username [github-actions]: " SERVER_USER
SERVER_USER=${SERVER_USER:-github-actions}
read -p "Enter path to SSH private key [~/.ssh/railzway-deploy]: " SSH_KEY
SSH_KEY=${SSH_KEY:-~/.ssh/railzway-deploy}
SSH_KEY="${SSH_KEY/#\~/$HOME}"

# Validate SSH Key
if [ ! -f "$SSH_KEY" ]; then
  echo -e "${RED}❌ Error: SSH key not found: $SSH_KEY${NC}"
  exit 1
fi

# Load .env locally to pre-fill prompts (optional)
if [ -f .env.production ]; then
  source .env.production
fi

echo ""
echo "Enter PostgreSQL Credentials (for the instance RUNNING ON NOMAD):"
read -p "Admin Host [${DB_HOST:-10.148.0.2}]: " ADMIN_HOST
read -p "Admin User [postgres]: " ADMIN_USER
ADMIN_USER=${ADMIN_USER:-postgres}
read -s -p "Admin Password: " ADMIN_PASSWORD
echo ""

APP_USER=${DB_USER:-railzway}
APP_PASS=${DB_PASSWORD:-railzway-secret}
APP_DB=${DB_NAME:-railzway_auth}

echo ""
echo "----------------------------------------------"
echo "Target Server: $SERVER_IP"
echo "Will Create:"
echo "  User: $APP_USER"
echo "  DB:   $APP_DB"
echo "----------------------------------------------"
read -p "Run setup? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  exit 0
fi

echo ""
echo "📦 Connecting to server to find Postgres container..."

# We execute the logic remotely via SSH
ssh -i "$SSH_KEY" ${SERVER_USER}@${SERVER_IP} "bash -s" <<EOF
  set -e
  
  # Find Postgres Container ID
  # We look for a container image containing 'postgres' that is running
  echo "🔍 Searching for running Postgres container..."
  CONTAINER_ID=\$(docker ps --format '{{.ID}} {{.Image}}' | grep postgres | head -n 1 | awk '{print \$1}')
  
  if [ -z "\$CONTAINER_ID" ]; then
    echo "❌ Error: No Postgres container found running on this server!"
    echo "Running containers:"
    docker ps --format '{{.ID}} {{.Image}}'
    exit 1
  fi
  
  echo "✅ Found Postgres container: \$CONTAINER_ID"
  
  # Helper to run psql inside the container
  run_psql() {
    docker exec -e PGPASSWORD="$ADMIN_PASSWORD" \$CONTAINER_ID psql -U "$ADMIN_USER" -d postgres -c "\$1"
  }
  
  echo ""
  echo "👤 1. Creating User '$APP_USER'..."
  # Use DO block to handle 'already exists' gracefully
  docker exec -e PGPASSWORD="$ADMIN_PASSWORD" \$CONTAINER_ID psql -U "$ADMIN_USER" -d postgres -c "
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$APP_USER') THEN
    CREATE ROLE $APP_USER LOGIN PASSWORD '$APP_PASS';
    RAISE NOTICE 'User $APP_USER created';
  ELSE
    RAISE NOTICE 'User $APP_USER already exists';
    ALTER ROLE $APP_USER WITH PASSWORD '$APP_PASS';
  END IF;
END
\$\$;"
  
  echo ""
  echo "🗄️  2. Creating Database '$APP_DB'..."
  # Check existence
  EXISTS=\$(docker exec -e PGPASSWORD="$ADMIN_PASSWORD" \$CONTAINER_ID psql -U "$ADMIN_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$APP_DB'")
  
  if [ "\$EXISTS" = "1" ]; then
    echo "   Database '$APP_DB' already exists."
  else
    run_psql "CREATE DATABASE $APP_DB OWNER $APP_USER;"
    echo "   Database '$APP_DB' created."
  fi
  
  echo ""
  echo "🔑 3. Granting Privileges..."
  run_psql "GRANT ALL PRIVILEGES ON DATABASE $APP_DB TO $APP_USER;"
  run_psql "GRANT ALL ON SCHEMA public TO $APP_USER;"
  
  echo ""
  echo "✅ Remote setup complete!"
EOF
