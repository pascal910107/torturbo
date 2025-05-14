#!/usr/bin/env bash
set -e

OS="$(uname -s)"

case "$OS" in
  Linux)
    echo "Detected Linux system, starting Tor installation..."
    if command -v apt-get &> /dev/null; then
      sudo apt-get update
      sudo apt-get install -y tor
    elif command -v dnf &> /dev/null; then
      sudo dnf install -y tor
    elif command -v pacman &> /dev/null; then
      sudo pacman -Sy --noconfirm tor
    else
      echo "Unsupported Linux distribution, please install Tor manually"
      exit 1
    fi
    ;;
  Darwin)
    echo "Detected macOS system, starting Tor installation..."
    if command -v brew &> /dev/null; then
      brew update
      brew install tor
    elif command -v port &> /dev/null; then
      sudo port selfupdate
      sudo port install tor
    else
      echo "Homebrew or MacPorts not found, please install one of them first"
      exit 1
    fi
    ;;
  *)
    echo "Not Linux or macOS, please use PowerShell script for Windows installation"
    exit 1
    ;;
esac

echo "=== Tor installation completed ==="
