#!/bin/bash
# Initial setup script for railzway-auth deployment
# This script is for ONE-TIME initial setup only
# After this, deployments will be handled by CI/CD (GitHub Actions)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🚀 Railzway Auth - Initial Server Setup"
echo "=========================================="
echo ""

# Configuration
read -p "Enter server IP [34.87.70.45]: " SERVER_IP
SERVER_IP=${SERVER_IP:-34.87.70.45}
read -p "Enter SSH username [github-actions]: " SERVER_USER
SERVER_USER=${SERVER_USER:-github-actions}
read -p "Enter path to SSH private key [~/.ssh/railzway-deploy]: " SSH_KEY
SSH_KEY=${SSH_KEY:-~/.ssh/railzway-deploy}

# Expand tilde
SSH_KEY="${SSH_KEY/#\~/$HOME}"

# Validate inputs
if [ -z "$SERVER_IP" ]; then
  echo -e "${RED}❌ Error: Server IP is required${NC}"
  exit 1
fi

if [ ! -f "$SSH_KEY" ]; then
  echo -e "${RED}❌ Error: SSH key not found: $SSH_KEY${NC}"
  exit 1
fi

ENV_FILE=".env.production"
if [ ! -f "$ENV_FILE" ]; then
  echo -e "${YELLOW}⚠ .env.production not found, checking .env...${NC}"
  ENV_FILE=".env"
  if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}❌ Error: .env.production or .env not found${NC}"
    echo "Please run this script from the project root directory"
    exit 1
  fi
fi

echo ""
echo "📋 Configuration:"
echo "  Server IP: $SERVER_IP"
echo "  SSH User: $SERVER_USER"
echo "  SSH Key: $SSH_KEY"
echo "  Env File: $ENV_FILE"
echo ""

read -p "Continue with setup? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Setup cancelled."
  exit 0
fi

echo ""
echo "📦 Step 1: Setting up directories on server..."
ssh -i "$SSH_KEY" ${SERVER_USER}@${SERVER_IP} << 'EOF'
  # Create directories with sudo
  # We use /opt/railzway/railzway-auth for better isolation from railzway-cloud
  sudo mkdir -p /opt/railzway/railzway-auth
  sudo chown -R ${USER}:${USER} /opt/railzway/railzway-auth
  
  # Set permissions
  sudo chmod 755 /opt/railzway/railzway-auth
EOF
echo -e "${GREEN}✓ Server directories created (/opt/railzway/railzway-auth)${NC}"

echo ""
echo "📦 Step 2: Copying environment file..."
scp -i "$SSH_KEY" "$ENV_FILE" ${SERVER_USER}@${SERVER_IP}:/opt/railzway/railzway-auth/.env.auth
echo -e "${GREEN}✓ Environment file copied to /opt/railzway/railzway-auth/.env${NC}"

echo ""
echo "📦 Step 3: Copying deployment scripts..."
scp -i "$SSH_KEY" \
  deployments/nomad/setup-consul-kv.sh \
  deployments/nomad/deploy.sh \
  deployments/nomad/railzway-auth.nomad \
  ${SERVER_USER}@${SERVER_IP}:/opt/railzway/railzway-auth/
echo -e "${GREEN}✓ Deployment scripts copied${NC}"

echo ""
echo "📦 Step 4: Finalizing server setup..."
ssh -i "$SSH_KEY" ${SERVER_USER}@${SERVER_IP} << 'EOF'
  cd /opt/railzway/railzway-auth
  
  # Make scripts executable
  chmod +x *.sh
  
  # Populate Consul KV
  echo "Populating Consul KV..."
  ./setup-consul-kv.sh .env.auth
  
  echo ""
  echo "✅ Files setup complete:"
  ls -lh /opt/railzway/railzway-auth/
EOF
echo -e "${GREEN}✓ Setup finalized${NC}"

echo ""
echo "=========================================="
echo -e "${GREEN}✅ Initial setup complete!${NC}"
echo ""
echo "📋 Next steps:"
echo ""
echo "1. Verify server setup:"
echo "   ssh -i $SSH_KEY ${SERVER_USER}@${SERVER_IP}"
echo "   ls -la /opt/railzway/railzway-auth"
echo "   consul kv get -recurse railzway-auth/"
echo ""
echo "2. Setup GitHub Secrets for CI/CD:"
echo "   Go to: https://github.com/railzwaylabs/railzway-auth/settings/secrets/actions"
echo "   Add/Update these secrets:"
echo "   - GCE_HOST_PROD_1 = $SERVER_IP"
echo "   - GCE_USERNAME_PROD_1 = $SERVER_USER"
echo "   - GCE_SSH_KEY_PROD_1 = (content of $SSH_KEY)"
echo "   - GITHUB_TOKEN = (already exists)"
echo "   - SLACK_WEBHOOK_URL = (for notifications)"
echo ""
echo "3. Deploy:"
echo "   ssh -i $SSH_KEY ${SERVER_USER}@${SERVER_IP}"
echo "   cd /opt/railzway/railzway-auth"
echo "   ./deploy.sh v0.0.1  # Or your desired version"
echo ""
