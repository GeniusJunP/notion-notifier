#!/usr/bin/env bash
set -euo pipefail

HOST="example-server"
REMOTE_BASE="~/.local/share/notion-notifier"
MODE="install"

usage() {
  cat <<'USAGE'
usage: deploy-mac.sh [-h host] [-m install|update]

  -h host   SSH host (default: example-server)
  -m mode   install: full install (binary + config/env/db + service setup)
            update : binary-only rollout (switch release + restart service)
USAGE
}

while getopts "h:m:" opt; do
  case "$opt" in
    h) HOST="$OPTARG" ;;
    m) MODE="$OPTARG" ;;
    *)
      usage
      exit 1
      ;;
  esac
done

if [[ "$MODE" != "install" && "$MODE" != "update" ]]; then
  echo "invalid mode: $MODE"
  usage
  exit 1
fi

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
if [[ "$MODE" == "install" ]]; then
  ssh "$HOST" "mkdir -p $REMOTE_BASE/releases/$RELEASE_TS $REMOTE_BASE/shared ~/.config/systemd/user"
else
  ssh "$HOST" "mkdir -p $REMOTE_BASE/releases/$RELEASE_TS"
fi

echo "[4/6] Upload artifacts and runtime files"
scp "$ROOT_DIR/build/notion-notifier" "$HOST:$REMOTE_BASE/releases/$RELEASE_TS/notion-notifier"
if [[ "$MODE" == "install" ]]; then
  scp "$ROOT_DIR/config.yaml" "$HOST:$REMOTE_BASE/shared/config.yaml"
  scp "$ROOT_DIR/env.yaml" "$HOST:$REMOTE_BASE/shared/env.yaml"
  scp "$ROOT_DIR/data.db" "$HOST:$REMOTE_BASE/shared/data.db"
fi

echo "[5/6] Set permissions and switch release"
if [[ "$MODE" == "install" ]]; then
  ssh "$HOST" "
  chmod 700 $REMOTE_BASE $REMOTE_BASE/shared
  chmod 755 $REMOTE_BASE/releases/$RELEASE_TS $REMOTE_BASE/releases/$RELEASE_TS/notion-notifier
  chmod 600 $REMOTE_BASE/shared/config.yaml $REMOTE_BASE/shared/env.yaml $REMOTE_BASE/shared/data.db
  ln -sfn $REMOTE_BASE/releases/$RELEASE_TS $REMOTE_BASE/current
  "
else
  ssh "$HOST" "
  chmod 755 $REMOTE_BASE/releases/$RELEASE_TS $REMOTE_BASE/releases/$RELEASE_TS/notion-notifier
  ln -sfn $REMOTE_BASE/releases/$RELEASE_TS $REMOTE_BASE/current
  "
fi

echo "[6/6] ${MODE}: service operation"
if [[ "$MODE" == "install" ]]; then
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
else
  ssh "$HOST" '
systemctl --user restart notion-notifier.service
systemctl --user --no-pager --full status notion-notifier.service | sed -n "1,20p"
'
fi

echo "Deployment completed ($MODE): $RELEASE_TS"
