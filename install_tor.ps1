<#
.SYNOPSIS
  Install Tor Expert Bundle in the project's third_party\tor folder (no system-level installation, no Chocolatey required)
.DESCRIPTION
  This script automatically downloads the latest Windows x86_64 version of Tor Expert Bundle,
  extracts it to the third_party\tor folder in the same directory as the script, and cleans up temporary files.
#>

param (
    [string]$ProjectRoot = (Split-Path -Parent $MyInvocation.MyCommand.Definition),
    [string]$TorVersion = "14.0.9"
)

# 1. Define download URL and destination folder
$torFileName = "tor-expert-bundle-windows-x86_64-$TorVersion.tar.gz"
$torUrl = "https://archive.torproject.org/tor-package-archive/torbrowser/$TorVersion/$torFileName"
$destDir = Join-Path $ProjectRoot "third_party\tor"
$tempArchive = Join-Path $env:TEMP $torFileName

# 2. Create folder
if (Test-Path $destDir) {
    Write-Host "Detected existing $destDir, will remove old files..."
    Remove-Item -Recurse -Force $destDir
}
Write-Host "Creating folder: $destDir"
New-Item -ItemType Directory -Path $destDir | Out-Null

# 3. Download Tor archive
Write-Host "Downloading Tor from $torUrl..."
Invoke-WebRequest -Uri $torUrl -OutFile $tempArchive -UseBasicParsing

# 4. Extract .tar.gz (requires Windows 10+ built-in tar support)
Write-Host "Extracting $tempArchive to $destDir..."
# tar.exe is usually built into Windows
& tar.exe -xzf $tempArchive -C $destDir

# 5. Clean up temporary files
Write-Host "Removing temporary file: $tempArchive"
Remove-Item $tempArchive -Force

# 6. Completion message
Write-Host "=== Tor has been installed to $destDir ==="
Write-Host "Example executable path:" (Join-Path $destDir "tor.exe")
