# Environment Setup Script for Windows PowerShell
# Run this script to set up your environment variables for the crypto trading bot

Write-Host "=== Crypto Trading Bot Environment Setup ===" -ForegroundColor Green
Write-Host ""

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "✓ Go is installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go is not installed. Please install Go 1.21 or higher from https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Setting up Kraken API credentials..." -ForegroundColor Yellow
Write-Host ""

# Get API credentials from user
$apiKey = Read-Host "Enter your Kraken API Key"
$privateKey = Read-Host "Enter your Kraken Private Key" -AsSecureString

# Convert secure string back to plain text for environment variable
$privateKeyPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($privateKey))

# Set environment variables for current session
$env:KRAKEN_API_KEY = $apiKey
$env:KRAKEN_PRIVATE_KEY = $privateKeyPlain

# Set permanent environment variables
Write-Host "Setting permanent environment variables..." -ForegroundColor Yellow
try {
    [Environment]::SetEnvironmentVariable("KRAKEN_API_KEY", $apiKey, "User")
    [Environment]::SetEnvironmentVariable("KRAKEN_PRIVATE_KEY", $privateKeyPlain, "User")
    Write-Host "✓ Environment variables set successfully" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to set permanent environment variables" -ForegroundColor Red
    Write-Host "You can set them manually or run this script as administrator" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Installing Go dependencies..." -ForegroundColor Yellow

# Initialize go module and install dependencies
go mod tidy

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Dependencies installed successfully" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to install dependencies" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Setup Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Review and customize config.json for your risk preferences"
Write-Host "2. Test with paper trading first"
Write-Host "3. Run the bot with: go run main.go"
Write-Host ""
Write-Host "⚠️  WARNING: Start with small amounts and monitor carefully!" -ForegroundColor Red
