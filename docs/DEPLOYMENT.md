# Deploy Todo Agent Backend ke AWS EC2

## 🚀 **Step-by-Step Deployment Guide**

### **Prerequisites**

- AWS Account dengan EC2 access
- SSH Key Pair untuk akses EC2
- Domain/subdomain (opsional)
- Gemini API Key
- Supabase project setup

---

## 📋 **Step 1: Setup AWS EC2 Instance**

### **1.1 Launch EC2 Instance**

```bash
# Login ke AWS Console
# Navigate ke EC2 Dashboard
# Click "Launch Instance"

# Instance Configuration:
- Name: todo-agent-backend
- AMI: Amazon Linux 2023 (free tier eligible)
- Instance Type: t2.micro (1 vCPU, 1 GB RAM)
- Key Pair: Create new atau pilih existing
- Security Group: Create new dengan rules:
  - SSH (22): Your IP
  - HTTP (80): 0.0.0.0/0
  - HTTPS (443): 0.0.0.0/0
  - Custom TCP (8080): 0.0.0.0/0
- Storage: 8 GB gp3 (free tier)
```

### **1.2 Connect ke Instance**

```bash
# Download key pair (.pem file)
chmod 400 your-key.pem

# Connect via SSH
ssh -i your-key.pem ec2-user@your-ec2-public-ip
```

---

## ⚙️ **Step 2: Setup Server Environment**

### **2.1 Update System & Install Dependencies**

```bash
# Update system
sudo yum update -y

# Install Docker
sudo yum install -y docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -a -G docker ec2-user

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Install Git
sudo yum install -y git

# Install Go (jika ingin build di server)
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install curl dan jq untuk testing
sudo yum install -y curl jq

# Reboot untuk apply group changes
sudo reboot
```

---

## 📦 **Step 3: Deploy Application**

### **Option A: Deploy dengan Docker (Recommended)**

#### **3.1 Clone Repository**

```bash
# SSH ke server lagi setelah reboot
ssh -i your-key.pem ec2-user@your-ec2-public-ip

# Clone repository
git clone https://github.com/your-username/todo-agent-backend.git
cd todo-agent-backend
```

#### **3.2 Setup Environment**

```bash
# Copy config file
cp config/config.example.yaml config/config.yaml

# Edit config dengan environment variables
nano config/config.yaml
```

Edit file config:

```yaml
server:
  port: 8080
  mode: release # Ubah ke release untuk production
  api_key: "${API_KEY}"
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120
  max_file_size: 5242880

gemini:
  api_key: "${GEMINI_API_KEY}"
  model: "gemini-1.5-flash"
  timeout: 30
  max_retries: 3

supabase:
  url: "${SUPABASE_URL}"
  key: "${SUPABASE_KEY}"
  timeout: 30
  max_retries: 3
# ... rest of config
```

#### **3.3 Create Environment File**

```bash
# Create .env file
nano .env
```

Isi file `.env`:

```bash
API_KEY=your-secure-api-key-here
GEMINI_API_KEY=your-gemini-api-key
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-supabase-anon-key
```

#### **3.4 Deploy dengan Docker Compose**

```bash
# Build dan start services
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f
```

### **Option B: Deploy sebagai Binary Systemd Service**

#### **3.1 Build Application**

```bash
# Clone dan build
git clone https://github.com/your-username/todo-agent-backend.git
cd todo-agent-backend

# Install dependencies
go mod tidy

# Build binary
make build

# Or manual build
CGO_ENABLED=0 GOOS=linux go build -o bin/todo-agent cmd/server/main.go
```

#### **3.2 Run Deployment Script**

```bash
# Make script executable
chmod +x deploy/deploy.sh

# Run deployment script
sudo ./deploy/deploy.sh
```

Script akan:

- Create user `todo-agent`
- Setup directory `/opt/todo-agent`
- Install systemd service
- Configure log rotation
- Setup swap file
- Start service

#### **3.3 Configure Environment**

```bash
# Edit environment file
sudo nano /opt/todo-agent/.env
```

Isi dengan:

```bash
GEMINI_API_KEY=your-gemini-api-key
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-supabase-anon-key
```

#### **3.4 Start Service**

```bash
# Restart service dengan config baru
sudo systemctl restart todo-agent

# Check status
sudo systemctl status todo-agent

# View logs
sudo journalctl -u todo-agent -f
```

---

## 🔒 **Step 4: Setup Reverse Proxy (Nginx)**

### **4.1 Install Nginx**

```bash
sudo yum install -y nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

### **4.2 Configure Nginx**

```bash
sudo nano /etc/nginx/conf.d/todo-agent.conf
```

Isi config:

```nginx
upstream todo_agent {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name your-domain.com;  # Ganti dengan domain Anda

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # File upload size limit
    client_max_body_size 5M;

    location / {
        proxy_pass http://todo_agent;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint
    location /healthz {
        proxy_pass http://todo_agent;
        access_log off;
    }
}
```

### **4.3 Test & Restart Nginx**

```bash
# Test configuration
sudo nginx -t

# Restart nginx
sudo systemctl restart nginx
```

---

## 🔐 **Step 5: Setup SSL dengan Let's Encrypt (Optional)**

### **5.1 Install Certbot**

```bash
sudo yum install -y python3-pip
sudo pip3 install certbot certbot-nginx
```

### **5.2 Obtain SSL Certificate**

```bash
# Get certificate (ganti dengan domain Anda)
sudo certbot --nginx -d your-domain.com

# Auto-renewal setup
sudo crontab -e
# Add line:
0 12 * * * /usr/bin/certbot renew --quiet
```

---

## 📊 **Step 6: Monitoring & Maintenance**

### **6.1 Setup Log Monitoring**

```bash
# View application logs
sudo journalctl -u todo-agent -f

# View nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log

# View system resources
htop
free -h
df -h
```

### **6.2 Setup CloudWatch Monitoring**

```bash
# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
sudo rpm -U ./amazon-cloudwatch-agent.rpm

# Configure CloudWatch (perlu IAM role dengan CloudWatch permissions)
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-config-wizard
```

### **6.3 Backup & Updates**

```bash
# Create backup script
sudo nano /opt/scripts/backup.sh
```

Isi script:

```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/backups"

mkdir -p $BACKUP_DIR

# Backup application
tar -czf $BACKUP_DIR/todo-agent-$DATE.tar.gz /opt/todo-agent

# Backup logs
tar -czf $BACKUP_DIR/logs-$DATE.tar.gz /var/log/nginx /opt/todo-agent/logs

# Keep only last 7 days
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "Backup completed: $DATE"
```

```bash
# Make executable
sudo chmod +x /opt/scripts/backup.sh

# Add to crontab (daily backup)
sudo crontab -e
# Add line:
0 2 * * * /opt/scripts/backup.sh
```

---

## 🧪 **Step 7: Testing Deployment**

### **7.1 Health Check**

```bash
# Direct test
curl http://your-ec2-ip:8080/healthz

# Through nginx
curl http://your-domain.com/healthz
```

### **7.2 API Testing**

```bash
# Test process endpoint
curl -X POST "http://your-domain.com/process" \
  -H "X-API-Key: your-api-key" \
  -F "type=text" \
  -F "content=Meeting tomorrow at 10am, review code" \
  -F "user_id=test-user"

# Test status endpoint
curl -X GET "http://your-domain.com/status/job-id" \
  -H "X-API-Key: your-api-key"
```

### **7.3 Load Testing**

```bash
# Install Apache Bench
sudo yum install -y httpd-tools

# Simple load test
ab -n 100 -c 10 http://your-domain.com/healthz
```

---

## 🔧 **Troubleshooting**

### **Common Issues:**

#### **Service tidak start:**

```bash
# Check service status
sudo systemctl status todo-agent

# Check logs
sudo journalctl -u todo-agent -n 50

# Check config
sudo /opt/todo-agent/bin/todo-agent --help
```

#### **Port tidak bisa diakses:**

```bash
# Check firewall
sudo iptables -L

# Check security group di AWS Console
# Ensure port 8080 dan 80 terbuka

# Check process listening
sudo netstat -tlnp | grep :8080
```

#### **Memory issues:**

```bash
# Check memory usage
free -h

# Check swap
swapon -s

# Restart service jika perlu
sudo systemctl restart todo-agent
```

---

## 📈 **Optimization untuk Production**

### **Performance Tuning:**

```bash
# Increase file descriptor limits
echo '* soft nofile 65536' | sudo tee -a /etc/security/limits.conf
echo '* hard nofile 65536' | sudo tee -a /etc/security/limits.conf

# Optimize kernel parameters
echo 'net.core.somaxconn = 1024' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 1024' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

### **Auto-scaling Setup:**

- Setup Auto Scaling Group
- Configure Application Load Balancer
- Setup CloudWatch alarms
- Create AMI dari configured instance

---

## ✅ **Deployment Checklist**

- [ ] EC2 instance launched
- [ ] Security groups configured
- [ ] SSH access working
- [ ] Dependencies installed
- [ ] Application deployed
- [ ] Environment variables set
- [ ] Service running
- [ ] Nginx configured
- [ ] SSL certificate installed (if using domain)
- [ ] Monitoring setup
- [ ] Backup configured
- [ ] Health check passing
- [ ] API endpoints tested
- [ ] Load testing completed

**🎉 Deployment Complete!**

Aplikasi Anda sekarang berjalan di: `http://your-domain.com` atau `http://your-ec2-ip:8080`
