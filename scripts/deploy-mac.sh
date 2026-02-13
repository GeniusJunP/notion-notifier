#!/usr/bin/env bash
set -euo pipefail

HOST="example-server"
REMOTE_BASE="~/.local/share/notion-notifier"

while getopts "h:" opt; do
  case "$opt" in
    h) HOST="$OPTARG" ;;
    *)
      echo "usage: $0 [-h host]"
      exit 1
      ;;
  esac
done

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
RELEASE_TS="$(date +%Y%m%d%H%M%S)"

echo "[1/6] Build web assets"
cd "$ROOT_DIR/web"
npm ci
npm run build

echo "[2/6] Build linux/arm64 binary"
cd "$ROOT_DIR"
mkdir -p build
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o build/notion-notifier ./cmd/notion-notifier

echo "[3/6] Prepare remote directories"
ssh "$HOST" "mkdir -p $REMOTE_BASE/releases/$RELEASE_TS $REMOTE_BASE/shared ~/.config/systemd/user"

echo "[4/6] Upload artifacts and runtime files"
scp "$ROOT_DIR/build/notion-notifier" "$HOST:$REMOTE_BASE/releases/$RELEASE_TS/notion-notifier"
scp "$ROOT_DIR/config.yaml" "$HOST:$REMOTE_BASE/shared/config.yaml"
scp "$ROOT_DIR/env.yaml" "$HOST:$REMOTE_BASE/shared/env.yaml"
scp "$ROOT_DIR/data.db" "$HOST:$REMOTE_BASE/shared/data.db"

echo "[5/6] Set permissions and switch release"
ssh "$HOST" "
chmod 700 $REMOTE_BASE $REMOTE_BASE/shared
chmod 755 $REMOTE_BASE/releases/$RELEASE_TS $REMOTE_BASE/releases/$RELEASE_TS/notion-notifier
chmod 600 $REMOTE_BASE/shared/config.yaml $REMOTE_BASE/shared/env.yaml $REMOTE_BASE/shared/data.db
ln -sfn $REMOTE_BASE/releases/$RELEASE_TS $REMOTE_BASE/current
"

echo "[6/6] Install/restart user service"
ssh "$HOST" '
cat > ~/.config/systemd/user/notion-notifier.service <<"UNIT"
[Unit]
Description=Notion Notifier
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
WorkingDirectory=%h/.local/share/notion-notifier/current
ExecStart=%h/.local/share/notion-notifier/current/notion-notifier -config %h/.local/share/notion-notifier/shared/config.yaml -env %h/.local/share/notion-notifier/shared/env.yaml -db %h/.local/share/notion-notifier/shared/data.db
Restart=always
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ReadWritePaths=%h/.local/share/notion-notifier

[Install]
WantedBy=default.target
UNIT

systemctl --user daemon-reload
systemctl --user enable --now notion-notifier.service
systemctl --user restart notion-notifier.service
loginctl show-user "$(id -un)" -p Linger | grep -q "Linger=yes" || loginctl enable-linger "$(id -un)"
systemctl --user --no-pager --full status notion-notifier.service | sed -n "1,20p"
'

echo "Deployment completed: $RELEASE_TS"
