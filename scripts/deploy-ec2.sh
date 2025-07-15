#!/bin/bash

# Todo Agent Backend - Quick EC2 Deployment Script
# Usage: ./deploy-ec2.sh <ec2-ip> <ssh-key-path>

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
EC2_IP="$1"
SSH_KEY="$2"
EC2_USER="ec2-user"
REMOTE_DIR="/opt/todo-agent"
SERVICE_NAME="todo-agent"

# Validation
if [ -z "$EC2_IP" ] || [ -z "$SSH_KEY" ]; then
    echo -e "${RED}Usage: $0 <ec2-ip> <ssh-key-path>${NC}"
    echo "Example: $0 54.123.45.67 ~/.ssh/my-key.pem"
    exit 1
fi

if [ ! -f "$SSH_KEY" ]; then
    echo -e "${RED}Error: SSH key file not found: $SSH_KEY${NC}"
    exit 1
fi

echo -e "${BLUE}🚀 Starting Todo Agent Backend Deployment to EC2${NC}"
echo -e "${BLUE}=================================================${NC}"
echo "Target: $EC2_USER@$EC2_IP"
echo "SSH Key: $SSH_KEY"
echo ""

# Function to run remote commands
run_remote() {
    ssh -i "$SSH_KEY" "$EC2_USER@$EC2_IP" "$@"
}

# Function to copy files
copy_to_remote() {
    scp -i "$SSH_KEY" -r "$1" "$EC2_USER@$EC2_IP:$2"
}

echo -e "${YELLOW}Step 1: Testing SSH connection...${NC}"
if run_remote "echo 'SSH connection successful'"; then
    echo -e "${GREEN}✓ SSH connection working${NC}"
else
    echo -e "${RED}✗ SSH connection failed${NC}"
    exit 1
fi

echo -e "${YELLOW}Step 2: Installing dependencies...${NC}"
run_remote "
    sudo yum update -y &&
    sudo yum install -y docker git curl jq &&
    sudo systemctl start docker &&
    sudo systemctl enable docker &&
    sudo usermod -a -G docker ec2-user
"
echo -e "${GREEN}✓ Dependencies installed${NC}"

echo -e "${YELLOW}Step 3: Installing Docker Compose...${NC}"
run_remote "
    sudo curl -L 'https://github.com/docker/compose/releases/latest/download/docker-compose-\$(uname -s)-\$(uname -m)' -o /usr/local/bin/docker-compose &&
    sudo chmod +x /usr/local/bin/docker-compose
"
echo -e "${GREEN}✓ Docker Compose installed${NC}"

echo -e "${YELLOW}Step 4: Building application locally...${NC}"
if [ ! -f "bin/todo-agent" ]; then
    echo "Building Go binary..."
    CGO_ENABLED=0 GOOS=linux go build -o bin/todo-agent cmd/server/main.go
fi
echo -e "${GREEN}✓ Application built${NC}"

echo -e "${YELLOW}Step 5: Creating remote directory...${NC}"
run_remote "sudo mkdir -p $REMOTE_DIR && sudo chown $EC2_USER:$EC2_USER $REMOTE_DIR"
echo -e "${GREEN}✓ Remote directory created${NC}"

echo -e "${YELLOW}Step 6: Copying application files...${NC}"
copy_to_remote "bin/todo-agent" "$REMOTE_DIR/"
copy_to_remote "config/" "$REMOTE_DIR/"
copy_to_remote "deploy/" "$REMOTE_DIR/"
copy_to_remote "docker-compose.yml" "$REMOTE_DIR/"
copy_to_remote "Dockerfile" "$REMOTE_DIR/"
echo -e "${GREEN}✓ Files copied${NC}"

echo -e "${YELLOW}Step 7: Setting up configuration...${NC}"
run_remote "
    cd $REMOTE_DIR &&
    cp config/config.example.yaml config/config.yaml &&
    touch .env
"

echo -e "${BLUE}Please configure your environment variables:${NC}"
echo "1. Gemini API Key"
echo "2. Supabase URL"
echo "3. Supabase Key"
echo "4. API Key for authentication"
echo ""

# Prompt for environment variables
read -p "Enter your Gemini API Key: " GEMINI_KEY
read -p "Enter your Supabase URL: " SUPABASE_URL
read -p "Enter your Supabase Key: " SUPABASE_KEY
read -p "Enter your API Key for authentication: " API_KEY

# Create environment file
cat > temp_env << EOF
GEMINI_API_KEY=$GEMINI_KEY
SUPABASE_URL=$SUPABASE_URL
SUPABASE_KEY=$SUPABASE_KEY
API_KEY=$API_KEY
EOF

copy_to_remote "temp_env" "$REMOTE_DIR/.env"
rm temp_env

echo -e "${GREEN}✓ Configuration set${NC}"

echo -e "${YELLOW}Step 8: Setting up systemd service...${NC}"
run_remote "
    sudo cp $REMOTE_DIR/deploy/todo-agent.service /etc/systemd/system/ &&
    sudo systemctl daemon-reload &&
    sudo systemctl enable $SERVICE_NAME
"
echo -e "${GREEN}✓ Systemd service configured${NC}"

echo -e "${YELLOW}Step 9: Setting up application user and permissions...${NC}"
run_remote "
    sudo useradd --system --home-dir $REMOTE_DIR --shell /bin/false todo-agent || true &&
    sudo chmod +x $REMOTE_DIR/todo-agent &&
    sudo chown -R todo-agent:todo-agent $REMOTE_DIR &&
    sudo mkdir -p /tmp/todo-agent &&
    sudo chown todo-agent:todo-agent /tmp/todo-agent
"
echo -e "${GREEN}✓ User and permissions set${NC}"

echo -e "${YELLOW}Step 10: Starting service...${NC}"
run_remote "sudo systemctl start $SERVICE_NAME"

# Wait a moment for service to start
sleep 3

echo -e "${YELLOW}Step 11: Checking service status...${NC}"
if run_remote "sudo systemctl is-active $SERVICE_NAME | grep -q active"; then
    echo -e "${GREEN}✓ Service is running${NC}"
else
    echo -e "${RED}✗ Service failed to start${NC}"
    echo "Checking logs..."
    run_remote "sudo journalctl -u $SERVICE_NAME -n 20"
    exit 1
fi

echo -e "${YELLOW}Step 12: Testing health check...${NC}"
sleep 2
if run_remote "curl -s http://localhost:8080/healthz | grep -q healthy"; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
    echo "Checking logs..."
    run_remote "sudo journalctl -u $SERVICE_NAME -n 10"
fi

echo -e "${YELLOW}Step 13: Setting up firewall (Security Groups)...${NC}"
echo -e "${BLUE}Please ensure the following ports are open in your EC2 Security Group:${NC}"
echo "- Port 22 (SSH): Your IP"
echo "- Port 8080 (App): 0.0.0.0/0"
echo "- Port 80 (HTTP): 0.0.0.0/0 (if using Nginx)"

echo ""
echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo -e "${BLUE}Your application is now running at:${NC}"
echo "- Health Check: http://$EC2_IP:8080/healthz"
echo "- API Endpoint: http://$EC2_IP:8080/process"
echo ""
echo -e "${BLUE}Useful commands:${NC}"
echo "- Check status: ssh -i $SSH_KEY $EC2_USER@$EC2_IP 'sudo systemctl status $SERVICE_NAME'"
echo "- View logs: ssh -i $SSH_KEY $EC2_USER@$EC2_IP 'sudo journalctl -u $SERVICE_NAME -f'"
echo "- Restart service: ssh -i $SSH_KEY $EC2_USER@$EC2_IP 'sudo systemctl restart $SERVICE_NAME'"
echo ""
echo -e "${BLUE}Test your API:${NC}"
echo "curl -X POST \"http://$EC2_IP:8080/process\" \\"
echo "  -H \"X-API-Key: $API_KEY\" \\"
echo "  -F \"type=text\" \\"
echo "  -F \"content=Meeting tomorrow at 10am\" \\"
echo "  -F \"user_id=test-user\""
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Setup domain name and SSL certificate"
echo "2. Configure Nginx reverse proxy"
echo "3. Setup monitoring and alerts"
echo "4. Configure backup strategy"
