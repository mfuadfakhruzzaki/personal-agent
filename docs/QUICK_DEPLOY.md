# 🚀 Quick Deployment Reference

## **TL;DR - Deploy ke EC2 dalam 5 Menit**

```bash
# 1. Launch EC2 t2.micro dengan security group port 8080 terbuka
# 2. Download SSH key (.pem file)
# 3. Run automated deployment script

./scripts/deploy-ec2.sh YOUR_EC2_IP PATH_TO_KEY.pem

# Script akan otomatis:
# ✅ Install dependencies (Docker, Git, etc)
# ✅ Build dan deploy aplikasi
# ✅ Setup systemd service
# ✅ Configure environment variables
# ✅ Start service dan test health check
```

---

## **Prerequisites (5 menit setup)**

### **1. AWS Setup**

```bash
# Launch EC2 Instance:
- AMI: Amazon Linux 2023
- Type: t2.micro (Free Tier)
- Security Group: SSH (22), HTTP (80), Custom (8080)
- Key Pair: Download .pem file
```

### **2. External Services**

```bash
# Gemini API (Google AI Studio):
# https://makersuite.google.com/app/apikey

# Supabase (Database):
# https://supabase.com/dashboard/new
# - Create new project
# - Get URL dan anon key
# - Run SQL schema (docs/DEPLOYMENT.md)
```

---

## **Deployment Options**

### **Option 1: Automated Script (Recommended)**

```bash
# Linux/Mac
chmod +x scripts/deploy-ec2.sh
./scripts/deploy-ec2.sh 54.123.45.67 ~/.ssh/my-key.pem

# Windows PowerShell
.\scripts\deploy-ec2.ps1 -EC2IP "54.123.45.67" -SSHKey "C:\path\to\key.pem"
```

### **Option 2: Docker Compose**

```bash
# SSH ke server
ssh -i your-key.pem ec2-user@your-ec2-ip

# Clone repository
git clone https://github.com/your-username/todo-agent-backend.git
cd todo-agent-backend

# Setup environment
cp config/config.example.yaml config/config.yaml
nano .env  # Set API keys

# Deploy
docker-compose up -d
```

### **Option 3: Systemd Service**

```bash
# Build dan deploy
make build
sudo ./deploy/deploy.sh

# Configure
sudo nano /opt/todo-agent/.env
sudo systemctl restart todo-agent
```

---

## **Post-Deployment Checklist**

### **1. Verify Deployment**

```bash
# Health check
curl http://YOUR_EC2_IP:8080/healthz

# API test
curl -X POST "http://YOUR_EC2_IP:8080/process" \
  -H "X-API-Key: YOUR_API_KEY" \
  -F "type=text" \
  -F "content=Test todo item" \
  -F "user_id=test-user"
```

### **2. Production Setup (Optional)**

```bash
# Setup domain & SSL
sudo yum install -y nginx certbot
sudo certbot --nginx -d your-domain.com

# Setup monitoring
# - CloudWatch agent
# - Log rotation
# - Backup scripts
```

---

## **Environment Variables**

```bash
# Required (set during deployment)
GEMINI_API_KEY=your-gemini-key
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-supabase-anon-key
API_KEY=your-secure-api-key

# Optional
GIN_MODE=release
CONFIG_PATH=/opt/todo-agent/config/config.yaml
```

---

## **Troubleshooting**

### **Service Issues**

```bash
# Check service status
sudo systemctl status todo-agent

# View logs
sudo journalctl -u todo-agent -f

# Restart service
sudo systemctl restart todo-agent
```

### **Network Issues**

```bash
# Check if port is listening
sudo netstat -tlnp | grep :8080

# Test from inside server
curl localhost:8080/healthz

# Check security group in AWS Console
# Ensure port 8080 is open to 0.0.0.0/0
```

### **Permission Issues**

```bash
# Fix ownership
sudo chown -R todo-agent:todo-agent /opt/todo-agent
sudo chmod +x /opt/todo-agent/todo-agent

# Fix temp directory
sudo mkdir -p /tmp/todo-agent
sudo chown todo-agent:todo-agent /tmp/todo-agent
```

---

## **Common Commands**

### **Service Management**

```bash
# Start/Stop/Restart
sudo systemctl start todo-agent
sudo systemctl stop todo-agent
sudo systemctl restart todo-agent

# Enable/Disable auto-start
sudo systemctl enable todo-agent
sudo systemctl disable todo-agent

# View status
sudo systemctl status todo-agent
```

### **Logs & Monitoring**

```bash
# Real-time logs
sudo journalctl -u todo-agent -f

# Last N lines
sudo journalctl -u todo-agent -n 50

# Logs since time
sudo journalctl -u todo-agent --since "1 hour ago"

# System resources
htop
free -h
df -h
```

### **Updates & Maintenance**

```bash
# Update application
cd /opt/todo-agent
git pull origin main
make build
sudo systemctl restart todo-agent

# Update config
sudo nano /opt/todo-agent/config/config.yaml
sudo systemctl restart todo-agent

# Backup
sudo tar -czf backup-$(date +%Y%m%d).tar.gz /opt/todo-agent
```

---

## **Quick Test Commands**

```bash
# Health check
curl http://localhost:8080/healthz

# Process text
curl -X POST "http://localhost:8080/process" \
  -H "X-API-Key: test-api-key" \
  -F "type=text" \
  -F "content=Meeting tomorrow at 10am, review code" \
  -F "user_id=test-user"

# Check job status (use job_id from above)
curl -X GET "http://localhost:8080/status/JOB_ID" \
  -H "X-API-Key: test-api-key"

# Load test
ab -n 100 -c 10 http://localhost:8080/healthz
```

---

## **Cost Optimization**

### **EC2 Free Tier Limits**

- **Instance Hours:** 750 hours/month (t2.micro)
- **Storage:** 30 GB EBS storage
- **Data Transfer:** 15 GB outbound
- **Monitoring:** Basic CloudWatch

### **Resource Usage**

- **Memory:** ~50MB idle, <400MB peak
- **CPU:** Low usage dengan async processing
- **Storage:** <1GB aplikasi + logs
- **Network:** Minimal dengan efficient API

---

## **Support & Documentation**

- **Full Deployment Guide:** `docs/DEPLOYMENT.md`
- **API Documentation:** `docs/API.md`
- **Implementation Details:** `IMPLEMENTATION.md`
- **Testing Scripts:** `scripts/test_api.*`

**🎉 Happy Deploying!**
