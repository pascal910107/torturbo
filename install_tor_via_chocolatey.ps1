<#
.SYNOPSIS
  Install Tor using Chocolatey (requires administrator privileges)
#>

# Check if running as administrator
$currentUser = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $currentUser.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
  Write-Error "Please run this script as administrator"
  exit 1
}

# Install Chocolatey (skip if already installed)
if (-not (Get-Command choco.exe -ErrorAction SilentlyContinue)) {
  Write-Host "Chocolatey not detected, starting installation..."
  Set-ExecutionPolicy Bypass -Scope Process -Force
  [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12
  iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
} else {
  Write-Host "Chocolatey is already installed"
}

# Install Tor
Write-Host "Starting Tor installation..."
choco install tor -y

Write-Host "=== Tor installation completed ==="
