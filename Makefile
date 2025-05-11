# Tor 版本設定
#    TOR_VERSION: 用於 Linux/macOS 的 Tor Browser 版本（tar.xz）
#    EXPERT_VERSION: 用於 Windows 的 Tor Expert Bundle 版本（tar.gz）
TOR_VERSION   := 13.5a1
EXPERT_VERSION:= 14.5.1

# ------------------------------------------------------------------
# OS 偵測：優先用環境變數 OS，再用 uname
ENV_OS  := $(OS)
UNAME_S := $(shell uname -s 2> /dev/null)
ifeq ($(ENV_OS),Windows_NT)
    DETECTED_OS := Windows_NT
else ifeq ($(UNAME_S),Linux)
    DETECTED_OS := Linux
else ifeq ($(UNAME_S),Darwin)
    DETECTED_OS := Darwin
else
    DETECTED_OS := Unknown
endif

# ------------------------------------------------------------------
# 路徑與命令設定
# Linux / macOS: 下載 torbrowser 目錄下的 .tar.xz
TOR_BASEURL := https://dist.torproject.org/torbrowser/$(TOR_VERSION)
TOR_TARBALL_LINUX  := tor-linux64-$(TOR_VERSION).tar.xz
TOR_TARBALL_DARWIN := tor-macos-$(TOR_VERSION).tar.xz

TOR_DOWNLOAD_LINUX  := curl -L $(TOR_BASEURL)/$(TOR_TARBALL_LINUX) -o $(TOR_TARBALL_LINUX)
TOR_EXTRACT_LINUX   := tar -xJf $(TOR_TARBALL_LINUX) -C third_party/tor --strip-components=1

TOR_DOWNLOAD_DARWIN := curl -L $(TOR_BASEURL)/$(TOR_TARBALL_DARWIN) -o $(TOR_TARBALL_DARWIN)
TOR_EXTRACT_DARWIN  := tar -xJf $(TOR_TARBALL_DARWIN) -C third_party/tor --strip-components=1

# Windows: 下載 Expert Bundle（包含 tor daemon、桥接、pluggable transports）
EXPERT_TARBALL := tor-expert-bundle-windows-x86_64-$(EXPERT_VERSION).tar.gz
EXPERT_URL     := https://archive.torproject.org/tor-package-archive/torbrowser/$(EXPERT_VERSION)/$(EXPERT_TARBALL)

# ------------------------------------------------------------------
# 目標定義
.PHONY: build ui bundle-tor release clean

# Cross-compile: linux, darwin, windows (amd64)
build:
ifeq ($(DETECTED_OS),Windows_NT)
	@echo "[*] building cross-platform binaries (Windows PowerShell)..."
	@powershell -NoProfile -Command " \
	  Set-Item Env:GOOS 'linux';    \
	  Set-Item Env:GOARCH 'amd64';  \
	  go mod tidy;                  \
	  go build -o dist/torturbo-linux-amd64   ./cmd/torturbo"
	@powershell -NoProfile -Command " \
	  Set-Item Env:GOOS 'darwin';   \
	  Set-Item Env:GOARCH 'amd64';  \
	  go mod tidy;                  \
	  go build -o dist/torturbo-darwin-amd64  ./cmd/torturbo"
	@powershell -NoProfile -Command " \
	  Set-Item Env:GOOS 'windows';  \
	  Set-Item Env:GOARCH 'amd64';  \
	  go mod tidy;                  \
	  go build -o dist/torturbo-windows-amd64.exe ./cmd/torturbo"
else
	@echo "[*] building cross-platform binaries..."
	@GOOS=linux   GOARCH=amd64 go mod tidy && GOOS=linux   GOARCH=amd64 go build -o dist/torturbo-linux-amd64   ./cmd/torturbo
	@GOOS=darwin  GOARCH=amd64 go mod tidy && GOOS=darwin  GOARCH=amd64 go build -o dist/torturbo-darwin-amd64  ./cmd/torturbo
	@GOOS=windows GOARCH=amd64 go mod tidy && GOOS=windows GOARCH=amd64 go build -o dist/torturbo-windows-amd64.exe ./cmd/torturbo
endif

# Next.js Dashboard
ui:
	@echo "[*] building Next.js dashboard..."
	@cd ui-src && npm install && npm run build

# 下載並解壓最新 Tor binary
bundle-tor:
	@echo "[*] Detected OS: $(DETECTED_OS)"
ifeq ($(DETECTED_OS),Windows_NT)
	@echo "[*] Preparing third_party/tor..."
	@powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path 'third_party/tor' | Out-Null"
	@echo "[*] Downloading Tor Expert Bundle: $(EXPERT_TARBALL)..."
	@curl -L $(EXPERT_URL) -o $(EXPERT_TARBALL)
	@echo "[*] Extracting to third_party/tor..."
	@tar -xzf $(EXPERT_TARBALL) -C third_party/tor --strip-components=1
	@powershell -NoProfile -Command "Remove-Item -LiteralPath '$(EXPERT_TARBALL)' -Force"
else ifeq ($(DETECTED_OS),Linux)
	@echo "[*] Preparing third_party/tor..."
	@mkdir -p third_party/tor
	@echo "[*] Downloading Tor Browser binary for Linux: $(TOR_TARBALL_LINUX)..."
	@$(TOR_DOWNLOAD_LINUX)
	@echo "[*] Extracting to third_party/tor..."
	@$(TOR_EXTRACT_LINUX)
	@rm -f $(TOR_TARBALL_LINUX)
else ifeq ($(DETECTED_OS),Darwin)
	@echo "[*] Preparing third_party/tor..."
	@mkdir -p third_party/tor
	@echo "[*] Downloading Tor Browser binary for macOS: $(TOR_TARBALL_DARWIN)..."
	@$(TOR_DOWNLOAD_DARWIN)
	@echo "[*] Extracting to third_party/tor..."
	@$(TOR_EXTRACT_DARWIN)
	@rm -f $(TOR_TARBALL_DARWIN)
else
	@echo "Unsupported OS: $(DETECTED_OS)"; exit 1
endif
	@echo "[OK] bundle-tor finished, contents in third_party/tor/"

# 一次性 Release（UI、Tor、二進位檔）
release: ui bundle-tor build
	@echo "[OK] full release complete."

# 清理
clean:
	@echo "[*] Cleaning artifacts..."
	@rm -rf dist internal/ui/static
	@rm -rf third_party/tor
	@rm -f tor-*.tar.xz tor-*.zip tor-expert-*.tar.gz
	@echo "[OK] clean finished."
