# Todo Agent Backend API Test Script (PowerShell)
# Make sure the server is running before executing this script

$BaseUrl = "http://localhost:8080"
$ApiKey = "your-api-key-here"

Write-Host "🚀 Testing Todo Agent Backend API" -ForegroundColor Green
Write-Host "=================================" -ForegroundColor Green

# Test 1: Health Check
Write-Host ""
Write-Host "1. Testing Health Check..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/healthz" -Method Get
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Process Text Input
Write-Host ""
Write-Host "2. Testing Text Processing..." -ForegroundColor Yellow

$form = @{
    type = 'text'
    content = 'Besok meeting dengan client jam 10, lalu review code, dan kirim laporan ke manager pada hari Jumat'
    user_id = 'test-user-123'
}

$headers = @{
    'X-API-Key' = $ApiKey
}

try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/process" -Method Post -Form $form -Headers $headers
    $response | ConvertTo-Json -Depth 3
    $jobId = $response.job_id
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Check Job Status
Write-Host ""
Write-Host "3. Checking Job Status..." -ForegroundColor Yellow
Start-Sleep -Seconds 2  # Wait a bit for processing

if ($jobId) {
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/status/$jobId" -Method Get -Headers $headers
        $response | ConvertTo-Json -Depth 5
    } catch {
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 4: Test Invalid API Key
Write-Host ""
Write-Host "4. Testing Invalid API Key..." -ForegroundColor Yellow

$invalidHeaders = @{
    'X-API-Key' = 'invalid-key'
}

try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/process" -Method Post -Form $form -Headers $invalidHeaders
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Expected error for invalid API key: $($_.Exception.Message)" -ForegroundColor Magenta
}

# Test 5: Test Missing Parameters
Write-Host ""
Write-Host "5. Testing Missing Parameters..." -ForegroundColor Yellow

$incompleteForm = @{
    type = 'text'
    user_id = 'test-user'
    # Missing content
}

try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/process" -Method Post -Form $incompleteForm -Headers $headers
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Expected error for missing parameters: $($_.Exception.Message)" -ForegroundColor Magenta
}

Write-Host ""
Write-Host "✅ API Testing Complete!" -ForegroundColor Green
