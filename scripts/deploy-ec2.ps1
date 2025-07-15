# Todo Agent Backend - EC2 Deployment Script (PowerShell)
# Usage: .\deploy-ec2.ps1 -EC2IP "your-ip" -SSHKey "path-to-key.pem"

param(
    [Parameter(Mandatory=$true)]
    [string]$EC2IP,
    
    [Parameter(Mandatory=$true)]
    [string]$SSHKey,
    
    [string]$EC2User = "ec2-user",
    [string]$RemoteDir = "/opt/todo-agent",
    [string]$ServiceName = "todo-agent"
)

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Cyan"
}

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Run-SSHCommand {
    param([string]$Command)
    $result = ssh -i $SSHKey "$EC2User@$EC2IP" $Command
    if ($LASTEXITCODE -ne 0) {
        throw "SSH command failed: $Command"
    }
    return $result
}

function Copy-ToRemote {
    param([string]$LocalPath, [string]$RemotePath)
    scp -i $SSHKey -r $LocalPath "$EC2User@$EC2IP`:$RemotePath"
    if ($LASTEXITCODE -ne 0) {
        throw "SCP failed: $LocalPath -> $RemotePath"
    }
}

# Validation
if (-not (Test-Path $SSHKey)) {
    Write-ColorOutput "Error: SSH key file not found: $SSHKey" $Colors.Red
    exit 1
}

Write-ColorOutput "🚀 Starting Todo Agent Backend Deployment to EC2" $Colors.Blue
Write-ColorOutput "=================================================" $Colors.Blue
Write-ColorOutput "Target: $EC2User@$EC2IP" "White"
Write-ColorOutput "SSH Key: $SSHKey" "White"
Write-ColorOutput ""

try {
    Write-ColorOutput "Step 1: Testing SSH connection..." $Colors.Yellow
    Run-SSHCommand "echo 'SSH connection successful'" | Out-Null
    Write-ColorOutput "✓ SSH connection working" $Colors.Green

    Write-ColorOutput "Step 2: Installing dependencies..." $Colors.Yellow
    Run-SSHCommand @"
        sudo yum update -y &&
        sudo yum install -y docker git curl jq &&
        sudo systemctl start docker &&
        sudo systemctl enable docker &&
        sudo usermod -a -G docker ec2-user
"@ | Out-Null
    Write-ColorOutput "✓ Dependencies installed" $Colors.Green

    Write-ColorOutput "Step 3: Installing Docker Compose..." $Colors.Yellow
    Run-SSHCommand @"
        sudo curl -L 'https://github.com/docker/compose/releases/latest/download/docker-compose-`$(uname -s)-`$(uname -m)' -o /usr/local/bin/docker-compose &&
        sudo chmod +x /usr/local/bin/docker-compose
"@ | Out-Null
    Write-ColorOutput "✓ Docker Compose installed" $Colors.Green

    Write-ColorOutput "Step 4: Building application locally..." $Colors.Yellow
    if (-not (Test-Path "bin\todo-agent.exe")) {
        Write-Host "Building Go binary..."
        $env:CGO_ENABLED = "0"
        $env:GOOS = "linux"
        go build -o bin/todo-agent cmd/server/main.go
    }
    Write-ColorOutput "✓ Application built" $Colors.Green

    Write-ColorOutput "Step 5: Creating remote directory..." $Colors.Yellow
    Run-SSHCommand "sudo mkdir -p $RemoteDir && sudo chown $EC2User`:$EC2User $RemoteDir" | Out-Null
    Write-ColorOutput "✓ Remote directory created" $Colors.Green

    Write-ColorOutput "Step 6: Copying application files..." $Colors.Yellow
    Copy-ToRemote "bin/todo-agent" "$RemoteDir/"
    Copy-ToRemote "config/" "$RemoteDir/"
    Copy-ToRemote "deploy/" "$RemoteDir/"
    Copy-ToRemote "docker-compose.yml" "$RemoteDir/"
    Copy-ToRemote "Dockerfile" "$RemoteDir/"
    Write-ColorOutput "✓ Files copied" $Colors.Green

    Write-ColorOutput "Step 7: Setting up configuration..." $Colors.Yellow
    Run-SSHCommand @"
        cd $RemoteDir &&
        cp config/config.example.yaml config/config.yaml &&
        touch .env
"@ | Out-Null

    Write-ColorOutput "Please configure your environment variables:" $Colors.Blue
    $GeminiKey = Read-Host "Enter your Gemini API Key"
    $SupabaseURL = Read-Host "Enter your Supabase URL"
    $SupabaseKey = Read-Host "Enter your Supabase Key"
    $ApiKey = Read-Host "Enter your API Key for authentication"

    # Create environment file
    $envContent = @"
GEMINI_API_KEY=$GeminiKey
SUPABASE_URL=$SupabaseURL
SUPABASE_KEY=$SupabaseKey
API_KEY=$ApiKey
"@
    $envContent | Out-File -FilePath "temp_env" -Encoding UTF8

    Copy-ToRemote "temp_env" "$RemoteDir/.env"
    Remove-Item "temp_env"
    Write-ColorOutput "✓ Configuration set" $Colors.Green

    Write-ColorOutput "Step 8: Setting up systemd service..." $Colors.Yellow
    Run-SSHCommand @"
        sudo cp $RemoteDir/deploy/todo-agent.service /etc/systemd/system/ &&
        sudo systemctl daemon-reload &&
        sudo systemctl enable $ServiceName
"@ | Out-Null
    Write-ColorOutput "✓ Systemd service configured" $Colors.Green

    Write-ColorOutput "Step 9: Setting up application user and permissions..." $Colors.Yellow
    Run-SSHCommand @"
        sudo useradd --system --home-dir $RemoteDir --shell /bin/false todo-agent `|| true &&
        sudo chmod +x $RemoteDir/todo-agent &&
        sudo chown -R todo-agent:todo-agent $RemoteDir &&
        sudo mkdir -p /tmp/todo-agent &&
        sudo chown todo-agent:todo-agent /tmp/todo-agent
"@ | Out-Null
    Write-ColorOutput "✓ User and permissions set" $Colors.Green

    Write-ColorOutput "Step 10: Starting service..." $Colors.Yellow
    Run-SSHCommand "sudo systemctl start $ServiceName" | Out-Null
    Start-Sleep -Seconds 3

    Write-ColorOutput "Step 11: Checking service status..." $Colors.Yellow
    try {
        $serviceStatus = Run-SSHCommand "sudo systemctl is-active $ServiceName"
        if ($serviceStatus -match "active") {
            Write-ColorOutput "✓ Service is running" $Colors.Green
        } else {
            throw "Service is not active"
        }
    } catch {
        Write-ColorOutput "✗ Service failed to start" $Colors.Red
        Write-ColorOutput "Checking logs..." "White"
        Run-SSHCommand "sudo journalctl -u $ServiceName -n 20"
        exit 1
    }

    Write-ColorOutput "Step 12: Testing health check..." $Colors.Yellow
    Start-Sleep -Seconds 2
    try {
        $healthCheck = Run-SSHCommand "curl -s http://localhost:8080/healthz"
        if ($healthCheck -match "healthy") {
            Write-ColorOutput "✓ Health check passed" $Colors.Green
        } else {
            throw "Health check failed"
        }
    } catch {
        Write-ColorOutput "✗ Health check failed" $Colors.Red
        Write-ColorOutput "Checking logs..." "White"
        Run-SSHCommand "sudo journalctl -u $ServiceName -n 10"
    }

    Write-ColorOutput "Step 13: Setting up firewall (Security Groups)..." $Colors.Yellow
    Write-ColorOutput "Please ensure the following ports are open in your EC2 Security Group:" $Colors.Blue
    Write-Host "- Port 22 (SSH): Your IP"
    Write-Host "- Port 8080 (App): 0.0.0.0/0"
    Write-Host "- Port 80 (HTTP): 0.0.0.0/0 (if using Nginx)"

    Write-Host ""
    Write-ColorOutput "🎉 Deployment completed successfully!" $Colors.Green
    Write-ColorOutput "================================================" $Colors.Green
    Write-Host ""
    Write-ColorOutput "Your application is now running at:" $Colors.Blue
    Write-Host "- Health Check: http://$EC2IP`:8080/healthz"
    Write-Host "- API Endpoint: http://$EC2IP`:8080/process"
    Write-Host ""
    Write-ColorOutput "Useful commands:" $Colors.Blue
    Write-Host "- Check status: ssh -i $SSHKey $EC2User@$EC2IP 'sudo systemctl status $ServiceName'"
    Write-Host "- View logs: ssh -i $SSHKey $EC2User@$EC2IP 'sudo journalctl -u $ServiceName -f'"
    Write-Host "- Restart service: ssh -i $SSHKey $EC2User@$EC2IP 'sudo systemctl restart $ServiceName'"
    Write-Host ""
    Write-ColorOutput "Test your API:" $Colors.Blue
    Write-Host @"
curl -X POST "http://$EC2IP`:8080/process" \
  -H "X-API-Key: $ApiKey" \
  -F "type=text" \
  -F "content=Meeting tomorrow at 10am" \
  -F "user_id=test-user"
"@
    Write-Host ""
    Write-ColorOutput "Next steps:" $Colors.Yellow
    Write-Host "1. Setup domain name and SSL certificate"
    Write-Host "2. Configure Nginx reverse proxy"
    Write-Host "3. Setup monitoring and alerts"
    Write-Host "4. Configure backup strategy"

} catch {
    Write-ColorOutput "❌ Deployment failed: $($_.Exception.Message)" $Colors.Red
    exit 1
}
