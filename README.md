# TorTurbo

> **多電路並行 + 快取 + 預建電路，讓 .onion 隱藏服務更快速、流暢。**  
> **Multiple Circuits + Cache + Pre-built Circuits for Faster .onion Hidden Services**

## 目錄 | Table of Contents

- [TorTurbo](#torturbo)
  - [目錄 | Table of Contents](#目錄--table-of-contents)
  - [簡介 | Introduction](#簡介--introduction)
  - [特色 | Features](#特色--features)
  - [效能提升原理 | Performance Improvement](#效能提升原理--performance-improvement)
  - [前置需求 | Prerequisites](#前置需求--prerequisites)
  - [安裝與編譯 | Installation \& Build](#安裝與編譯--installation--build)
  - [使用說明 | Usage](#使用說明--usage)
  - [設定與參數 | Configuration](#設定與參數--configuration)
  - [架構說明 | Architecture](#架構說明--architecture)
  - [專案結構 | Project Structure](#專案結構--project-structure)
  - [效能基準測試 | Benchmark](#效能基準測試--benchmark)
  - [授權 | License](#授權--license)

---

## 簡介 | Introduction

TorTurbo 是一款用於加速 Tor 隱藏服務（.onion）請求的本地代理工具，結合多電路並行、預建電路與本地靜態資源快取，能在不修改 Tor 核心協定的前提下，快速提升隱藏服務的載入速度。

TorTurbo is a local proxy tool designed to accelerate Tor hidden service (.onion) requests. By combining multiple circuits, pre-built circuits, and local static resource caching, it can significantly improve the loading speed of hidden services without modifying the Tor core protocol.

## 特色 | Features

* **多電路並行 | Multiple Circuits**: 同時維持多條 Tor 電路，依 RTT 排序分配 HTTP 請求，實現 2–3× 首訪加速。  
  Maintains multiple Tor circuits simultaneously, distributing HTTP requests based on RTT ranking, achieving 2-3× first-visit acceleration.

* **電路預建與心跳 | Circuit Pre-building & Heartbeat**: 啟動時自動建立並保持熱連線，省去 500–1000 ms 握手延遲。  
  Automatically establishes and maintains hot connections at startup, saving 500-1000 ms handshake delay.

* **本地快取 | Local Cache**: SQLite + 檔案系統儲存靜態資源，支援 ETag / If-Modified-Since，後續瀏覽秒開。  
  Uses SQLite + file system to store static resources, supports ETag / If-Modified-Since, enabling instant loading on subsequent visits.

* **Web 儀表板 | Web Dashboard**: Next.js + Tailwind 實時顯示各條電路 RTT、快取命中率等資訊。  
  Real-time display of circuit RTT, cache hit rates, and other metrics using Next.js + Tailwind.

* **跨平台可執行檔 | Cross-platform Executables**: 一鍵生成 Linux、macOS、Windows 執行檔，內置 Tor binary。  
  One-click generation of executables for Linux, macOS, and Windows, with built-in Tor binary.

## 效能提升原理 | Performance Improvement

1. **多電路並行 | Multiple Circuits**: 將 HTML、CSS、JS、圖片等資源分散到多條電路同時下載，平均吞吐量 ≈ 單電路×電路數。  
   Distributes HTML, CSS, JS, images, and other resources across multiple circuits for simultaneous download, with average throughput ≈ single circuit × number of circuits.

2. **預建電路 | Pre-built Circuits**: 客戶端背景自動建立 Circuit，並以心跳維持活性，減少每次連線握手開銷。  
   Automatically establishes circuits in the background and maintains activity through heartbeat, reducing connection handshake overhead.

3. **本地快取 | Local Cache**: 靜態資源採 TTL 緩存，動態檢測 ETag / Last-Modified，只回傳 304 或直接命中本地檔案。  
   Uses TTL caching for static resources, dynamically checks ETag / Last-Modified, returns 304 or directly serves local files.

## 前置需求 | Prerequisites

* Go 1.22+
* NPM & Node.js（用於建置前端儀表板 | for building frontend dashboard）
* `make` 工具 | `make` tool

## 安裝與編譯 | Installation & Build

```bash
# 取得專案 | Get the project
git clone https://github.com/pascal910107/torturbo.git
cd torturbo

# 建置前端儀表板 | Build frontend dashboard
make ui

# 下載並解壓 Tor binary | Download and extract Tor binary
make bundle-tor

# 交叉編譯三平臺執行檔 | Cross-compile for three platforms
make build
```

## 使用說明 | Usage

* 啟動代理與後端（Linux 範例）| Start proxy and backend (Linux example):

  ```bash
  # 監聽 8118 端口，並啟動 Web UI（預設 18000）| Listen on port 8118 and start Web UI (default 18000)
  ./dist/torturbo-linux-amd64 --listen 127.0.0.1:8118 --ui 127.0.0.1:18000
  ```
* 將瀏覽器 Proxy 設為 `127.0.0.1:8118`，即可透過 TorTurbo 訪問 .onion 網站。  
  Set browser proxy to `127.0.0.1:8118` to access .onion sites through TorTurbo.
* 打開瀏覽器至 `http://127.0.0.1:18000` 查看儀表板狀態。  
  Open browser to `http://127.0.0.1:18000` to view dashboard status.

## 設定與參數 | Configuration

| 參數<br>Parameter | 說明<br>Description | 預設值<br>Default |
| --- | --- | --- |
| `--listen` | HTTP 代理監聽位址<br>HTTP proxy listening address | `127.0.0.1:8118` |
| `--ui` | Web UI 監聽位址<br>Web UI listening address | `127.0.0.1:18000` |
| `--datadir` | 資料目錄（cache、Tor data 存放）<br>Data directory (cache, Tor data storage) | `$HOME/.torturbo` |
| `--v` | 啟用 DEBUG 日誌<br>Enable DEBUG logging | `false` |
| `CACHE_SIZE_LIMIT` | 本地快取上限（GB）<br>Local cache size limit (GB) | `2` |
| `CIRCUIT_NUM` | 同時建立的 Tor 電路數<br>Number of concurrent Tor circuits | `8` |

---

## 架構說明 | Architecture

```text
[Browser]
   ↓
[本地 TorTurbo Proxy | Local TorTurbo Proxy]
   ↓                          ↳ .onion Hidden Services
[Scheduler (多電路管理 | Multi-circuit Management)]  ←  [Tor Binary (內置 | Built-in)]
   ↓
[Cache (SQLite + FS)]
   ↓
[Next.js Web UI]
```

* **Proxy**: 解析請求、拆分多路、綁定自訂 HTTP Transport。  
  Parses requests, splits into multiple paths, binds custom HTTP Transport.

* **Scheduler**: 從 Controller 拿到 RTT 最小的 Circuit 來 Dial。  
  Gets the Circuit with the lowest RTT from Controller for Dial.

* **Controller & Circuit**: 利用 Bine library 啟動 Tor，動態維護多條 Circuit 及其 RTT。  
  Uses Bine library to start Tor, dynamically maintains multiple Circuits and their RTT.

* **Cache**: 先查本地，命中則直接回傳；未命中才走 Tor 並下載後存檔。  
  Checks local first, returns directly if hit; if miss, goes through Tor and saves after download.

## 專案結構 | Project Structure

```
torturbo/
├── cmd/torturbo          # 主程式入口 | Main program entry
├── internal/
│   ├── config           # 設定 | Configuration
│   ├── logger           # 日誌 | Logging
│   ├── cache            # 快取層 | Cache layer
│   ├── tunnel           # Tor Circuit 管理 | Tor Circuit management
│   ├── scheduler        # 請求排程 | Request scheduling
│   ├── proxy            # HTTP Proxy 實作 | HTTP Proxy implementation
│   └── ui               # Web UI 伺服器 | Web UI server
└── ui-src/               # Next.js 儀表板原始碼 | Next.js dashboard source code
```

## 效能基準測試 | Benchmark

| 測試項目<br>Test Item | 傳統 Tor (首訪)<br>Traditional Tor (First Visit) | TorTurbo (首訪)<br>TorTurbo (First Visit) | 加速比率<br>Speedup Ratio | 備註<br>Notes |
|:--------|:--------------|:---------------|:--------|:-----|
| 首頁載入時間<br>Homepage Load Time | 5–8 秒<br>5-8 seconds | 2–3 秒<br>2-3 seconds | ~2.5× | 包含 3–6 條並行電路與預建節點握手省時<br>Includes 3-6 parallel circuits and pre-built node handshake time saving |
| 重訪頁面<br>Revisit Page | 4–6 秒<br>4-6 seconds | <1 秒<br><1 second | ~5× | 靜態資源本地緩存命中<br>Static resource local cache hit |

## 授權 | License

本專案使用 **MIT License**，歡迎自由使用、修改及散佈。  
This project is licensed under the **MIT License**. Feel free to use, modify, and distribute.
