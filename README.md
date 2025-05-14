# TorTurbo

> **Multiple Circuits + Cache + Pre-built Circuits for Faster .onion Hidden Services**

## Table of Contents

- [TorTurbo](#torturbo)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Features](#features)
  - [Performance Improvement](#performance-improvement)
  - [Prerequisites](#prerequisites)
  - [Installation \& Build](#installation--build)
  - [Usage](#usage)
  - [Configuration](#configuration)
  - [Architecture](#architecture)
  - [Project Structure](#project-structure)
  - [Benchmark](#benchmark)

---

## Introduction

TorTurbo is a local proxy tool designed to accelerate Tor hidden service (.onion) requests. By combining multiple circuits, pre-built circuits, and local static resource caching, it can significantly improve the loading speed of hidden services without modifying the Tor core protocol.

## Features

* **Multiple Circuits**: Maintains multiple Tor circuits simultaneously, distributing HTTP requests based on RTT ranking, achieving 2-3× first-visit acceleration.

* **Circuit Pre-building & Heartbeat**: Automatically establishes and maintains hot connections at startup, saving 500-1000 ms handshake delay.

* **Local Cache**: Uses SQLite + file system to store static resources, supports ETag / If-Modified-Since, enabling instant loading on subsequent visits.

* **Web Dashboard**: Real-time display of circuit RTT, cache hit rates, and other metrics using Next.js + Tailwind.

* **Cross-platform Executables**: One-click generation of executables for Linux, macOS, and Windows, with built-in Tor binary.

## Performance Improvement

1. **Multiple Circuits**: Distributes HTML, CSS, JS, images, and other resources across multiple circuits for simultaneous download, with average throughput ≈ single circuit × number of circuits.

2. **Pre-built Circuits**: Automatically establishes circuits in the background and maintains activity through heartbeat, reducing connection handshake overhead.

3. **Local Cache**: Uses TTL caching for static resources, dynamically checks ETag / Last-Modified, returns 304 or directly serves local files.

## Prerequisites

* Go 1.22+
* NPM & Node.js (for building frontend dashboard)
* `make` tool

## Installation & Build

```bash
# Get the project
git clone https://github.com/pascal910107/torturbo.git
cd torturbo

# Build frontend dashboard
make ui

# Install Tor
# Windows (Choose one method):
# 1. Install to project directory (Recommended):
powershell -ExecutionPolicy Bypass -File install_tor.ps1
# 2. Install to system via Chocolatey (Requires admin privileges):
powershell -ExecutionPolicy Bypass -File install_tor_via_chocolatey.ps1

# Linux/macOS:
# Install to system (Requires admin privileges):
bash install_tor.sh

# Cross-compile for three platforms
make build
```

## Usage

* Start proxy and backend (Linux example):

  ```bash
  # Listen on port 8118 and start Web UI (default 18000)
  ./dist/torturbo-linux-amd64 --listen 127.0.0.1:8118 --ui 127.0.0.1:18000
  ```
* Set browser proxy to `127.0.0.1:8118` to access .onion sites through TorTurbo.
* Open browser to `http://127.0.0.1:18000` to view dashboard status.

## Configuration

| Parameter | Description | Default |
| --- | --- | --- |
| `--listen` | HTTP proxy listening address | `127.0.0.1:8118` |
| `--ui` | Web UI listening address | `127.0.0.1:18000` |
| `--datadir` | Data directory (cache, Tor data storage) | `$HOME/.torturbo` |
| `--torbundledir` | Path to bundled Tor directory | `third_party/tor` |
| `--v` | Enable DEBUG logging | `false` |
| `CACHE_SIZE_LIMIT` | Local cache size limit (GB) | `2` |
| `CIRCUIT_NUM` | Number of concurrent Tor circuits | `8` |

Usage Examples:
```bash
# Use custom data directory
./torturbo --datadir /path/to/data

# Use custom Tor directory
./torturbo --torbundledir /path/to/tor

# Enable debug logging
./torturbo --v

# Use multiple parameters
./torturbo --listen 0.0.0.0:8118 --ui 0.0.0.0:18000 --v
```

---

## Architecture

```text
[Browser]
   ↓
[Local TorTurbo Proxy]
   ↓                          ↳ .onion Hidden Services
[Scheduler (Multi-circuit Management)]  ←  [Tor Binary (Built-in)]
   ↓
[Cache (SQLite + FS)]
   ↓
[Next.js Web UI]
```

* **Proxy**: Parses requests, splits into multiple paths, binds custom HTTP Transport.

* **Scheduler**: Gets the Circuit with the lowest RTT from Controller for Dial.

* **Controller & Circuit**: Uses Bine library to start Tor, dynamically maintains multiple Circuits and their RTT.

* **Cache**: Checks local first, returns directly if hit; if miss, goes through Tor and saves after download.

## Project Structure

```
torturbo/
├── cmd/torturbo          # Main program entry
├── internal/
│   ├── config           # Configuration
│   ├── logger           # Logging
│   ├── cache            # Cache layer
│   ├── tunnel           # Tor Circuit management
│   ├── scheduler        # Request scheduling
│   ├── proxy            # HTTP Proxy implementation
│   └── ui               # Web UI server
└── ui-src/               # Next.js dashboard source code
```

## Benchmark

| Test Item | Traditional Tor (First Visit) | TorTurbo (First Visit) | Speedup Ratio | Notes |
|:--------|:--------------|:---------------|:--------|:-----|
| Homepage Load Time | 5-8 seconds | 2-3 seconds | ~2.5× | Includes 3-6 parallel circuits and pre-built node handshake time saving |
| Revisit Page | 4-6 seconds | <1 second | ~5× | Static resource local cache hit |
