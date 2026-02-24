#!/usr/bin/env bash
set -euo pipefail

# ==========================================
# Notion Notifier Installation Script
# ==========================================
# This script downloads the latest release from GitHub,
# sets up the XDG directories, and installs the systemd service.

REPO="YOUR_GITHUB_USERNAME/notion-notifier" # TODO: リポジトリの所有者に変更してください
BIN_NAME="notion-notifier"

echo "=> Installing $BIN_NAME from GitHub Releases ($REPO)..."

# 1. OSとアーキテクチャの判定
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [[ "$OS" != "linux" && "$OS" != "darwin" ]]; then
  echo "Error: This install script supports Linux and macOS (Darwin)."
  echo "For Windows, please run install.ps1 instead."
  exit 1
fi

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  DL_ARCH="amd64" ;;
  aarch64) DL_ARCH="arm64" ;;
  arm64)   DL_ARCH="arm64" ;;
  *)
    echo "Error: Unsupported architecture $ARCH"
    exit 1
    ;;
esac

ASSET_NAME="${BIN_NAME}-${OS}-${DL_ARCH}"

# 2. 最新リリースのダウンロードURLを取得
echo "=> Fetching latest release information..."
API_URL="https://api.github.com/repos/$REPO/releases/latest"

# GitHub REST APIからダウンロードURLを抽出 (jqがなくても動くようにgrepを使用)
DOWNLOAD_URL=$(curl -sSL "$API_URL" | grep "browser_download_url" | grep "$ASSET_NAME" | cut -d '"' -f 4 | head -n 1 || true)

if [[ -z "$DOWNLOAD_URL" ]]; then
  echo "Error: Could not find release asset '$ASSET_NAME'."
  echo "1. Please ensure $REPO is a public repository (or configure a PAT)."
  echo "2. Please ensure a release exists with the attached binary."
  exit 1
fi

# 3. ディレクトリの作成 (XDG Base Directory準拠)
XDG_BIN="$HOME/.local/bin"
XDG_CONFIG="$HOME/.config/$BIN_NAME"
XDG_DATA="$HOME/.local/share/$BIN_NAME"

echo "=> Creating directories..."
mkdir -p "$XDG_BIN" "$XDG_CONFIG" "$XDG_DATA" ~/.config/systemd/user

# 4. バイナリのダウンロードと配置
echo "=> Downloading $ASSET_NAME..."
TMP_BIN="/tmp/$BIN_NAME"
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_BIN"
chmod +x "$TMP_BIN"
mv "$TMP_BIN" "$XDG_BIN/$BIN_NAME"

# 5. 設定ファイル・DBの初期化 (既に存在する場合はスキップ)
if [[ ! -f "$XDG_CONFIG/config.yaml" ]]; then
  echo "=> Creating dummy config.yaml..."
  cat > "$XDG_CONFIG/config.yaml" <<EOF
server:
  port: 18080
auth:
  username: "admin"
  password: "password"
EOF
fi

if [[ ! -f "$XDG_CONFIG/env.yaml" ]]; then
  echo "=> Creating dummy env.yaml..."
  cat > "$XDG_CONFIG/env.yaml" <<EOF
notion_api_key: ""
database_id: ""
google_credentials_json: ""
EOF
fi

if [[ ! -f "$XDG_DATA/data.db" ]]; then
  echo "=> Initializing empty data.db..."
  touch "$XDG_DATA/data.db"
fi

chmod 600 "$XDG_CONFIG/config.yaml" "$XDG_CONFIG/env.yaml" "$XDG_DATA/data.db" 2>/dev/null || true

# 6. サービスの登録・自動起動設定
echo "=> Setting up background service..."

if [[ "$OS" == "linux" ]]; then
  # ======== Systemd (Linux) ========
  SERVICE_FILE="$HOME/.config/systemd/user/${BIN_NAME}.service"
  
  cat > "$SERVICE_FILE" <<UNIT
[Unit]
Description=Notion Notifier
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=%h/.local/share/${BIN_NAME}
ExecStart=%h/.local/bin/${BIN_NAME} -config %h/.config/${BIN_NAME}/config.yaml -env %h/.config/${BIN_NAME}/env.yaml -db %h/.local/share/${BIN_NAME}/data.db
Restart=always
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ReadWritePaths=%h/.local/share/${BIN_NAME} %h/.config/${BIN_NAME}

[Install]
WantedBy=default.target
UNIT

  systemctl --user daemon-reload
  systemctl --user enable --now ${BIN_NAME}.service

  # Lingerの有効化 (ユーザーがログアウトしてもサービスを継続動作させる)
  if ! loginctl show-user "$(id -un)" -p Linger | grep -q "Linger=yes"; then
    echo "=> Enabling systemd linger for $(id -un)..."
    loginctl enable-linger "$(id -un)"
  fi

  LOG_CMD="journalctl --user -u ${BIN_NAME}.service -f"

elif [[ "$OS" == "darwin" ]]; then
  # ======== Launchd (macOS) ========
  LAUNCH_AGENT_DIR="$HOME/Library/LaunchAgents"
  mkdir -p "$LAUNCH_AGENT_DIR"
  PLIST_FILE="$LAUNCH_AGENT_DIR/com.example.${BIN_NAME}.plist"
  
  cat > "$PLIST_FILE" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.example.${BIN_NAME}</string>
    <key>ProgramArguments</key>
    <array>
        <string>$XDG_BIN/$BIN_NAME</string>
        <string>-config</string>
        <string>$XDG_CONFIG/config.yaml</string>
        <string>-env</string>
        <string>$XDG_CONFIG/env.yaml</string>
        <string>-db</string>
        <string>$XDG_DATA/data.db</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>$XDG_DATA</string>
    <key>StandardOutPath</key>
    <string>$XDG_DATA/stdout.log</string>
    <key>StandardErrorPath</key>
    <string>$XDG_DATA/stderr.log</string>
</dict>
</plist>
PLIST

  launchctl unload "$PLIST_FILE" 2>/dev/null || true
  launchctl load -w "$PLIST_FILE"
  
  LOG_CMD="tail -f $XDG_DATA/stdout.log $XDG_DATA/stderr.log"
fi

echo ""
echo "==========================================================="
echo "✅ Installation complete!"
echo "The service has been started in the background."
echo "You can check the logs using:"
echo "  $LOG_CMD"
echo "==========================================================="
