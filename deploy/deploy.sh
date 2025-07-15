#!/bin/bash

# EC2 Deployment Script for Todo Agent Backend
# Usage: ./deploy.sh

set -e

# Configuration
APP_NAME="todo-agent"
APP_USER="todo-agent"
APP_DIR="/opt/$APP_NAME"
SERVICE_FILE="$APP_NAME.service"
BINARY_NAME="todo-agent"

echo "🚀 Starting deployment of $APP_NAME..."

# Create user if not exists
if ! id "$APP_USER" &>/dev/null; then
    echo "📝 Creating user $APP_USER..."
    sudo useradd --system --home-dir $APP_DIR --shell /bin/false $APP_USER
fi

# Create application directory
echo "📁 Setting up directories..."
sudo mkdir -p $APP_DIR/{bin,config,logs,tmp}
sudo chown -R $APP_USER:$APP_USER $APP_DIR

# Copy application files
echo "📦 Copying application files..."
sudo cp bin/$BINARY_NAME $APP_DIR/bin/
sudo cp config/config.example.yaml $APP_DIR/config/
sudo cp -r deploy $APP_DIR/

# Set permissions
sudo chmod +x $APP_DIR/bin/$BINARY_NAME
sudo chown -R $APP_USER:$APP_USER $APP_DIR

# Install systemd service
echo "⚙️ Installing systemd service..."
sudo cp $APP_DIR/deploy/$SERVICE_FILE /etc/systemd/system/
sudo systemctl daemon-reload

# Create environment file template
if [ ! -f "$APP_DIR/.env" ]; then
    echo "📋 Creating environment file template..."
    sudo tee $APP_DIR/.env > /dev/null <<EOF
GEMINI_API_KEY=your-gemini-api-key
SUPABASE_URL=your-supabase-url
SUPABASE_KEY=your-supabase-key
EOF
    sudo chown $APP_USER:$APP_USER $APP_DIR/.env
    sudo chmod 600 $APP_DIR/.env
    echo "⚠️  Please edit $APP_DIR/.env with your actual credentials"
fi

# Create config file if not exists
if [ ! -f "$APP_DIR/config/config.yaml" ]; then
    echo "📋 Creating config file..."
    sudo cp $APP_DIR/config/config.example.yaml $APP_DIR/config/config.yaml
    sudo chown $APP_USER:$APP_USER $APP_DIR/config/config.yaml
fi

# Setup log rotation
echo "📜 Setting up log rotation..."
sudo tee /etc/logrotate.d/$APP_NAME > /dev/null <<EOF
$APP_DIR/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 $APP_USER $APP_USER
    postrotate
        systemctl reload $APP_NAME || true
    endscript
}
EOF

# Setup swap file if not exists (for low memory environments)
if [ ! -f /swapfile ]; then
    echo "💾 Setting up swap file..."
    sudo fallocate -l 512M /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    
    # Add to fstab if not already there
    if ! grep -q "/swapfile" /etc/fstab; then
        echo "/swapfile none swap sw 0 0" | sudo tee -a /etc/fstab
    fi
fi

# Enable and start service
echo "🎯 Starting service..."
sudo systemctl enable $APP_NAME
sudo systemctl restart $APP_NAME

# Check service status
echo "🔍 Checking service status..."
sleep 2
sudo systemctl status $APP_NAME --no-pager

# Setup firewall if ufw is available
if command -v ufw &> /dev/null; then
    echo "🔒 Configuring firewall..."
    sudo ufw allow 8080/tcp comment "Todo Agent Backend"
fi

echo "✅ Deployment completed successfully!"
echo ""
echo "Next steps:"
echo "1. Edit $APP_DIR/.env with your API keys"
echo "2. Edit $APP_DIR/config/config.yaml if needed"
echo "3. Restart the service: sudo systemctl restart $APP_NAME"
echo "4. View logs: sudo journalctl -u $APP_NAME -f"
echo "5. Test the service: curl http://localhost:8080/healthz"
