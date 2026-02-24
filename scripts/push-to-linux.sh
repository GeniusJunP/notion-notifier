#!/usr/bin/env bash
set -euo pipefail

HOST="example-server"
REMOTE_BASE="~/.local/share/notion-notifier"
MODE="install"
TARGET_ARCH="arm64"

usage() {
  cat <<'USAGE'
usage: push-to-linux.sh [-h host] [-m install|update] [-a arch]

  -h host   SSH host (default: example-server)
  -m mode   install: full install (binary + config/env/db + service setup)
            update : binary-only rollout (switch release + restart service)
  -a arch   target architecture (default: arm64) (e.g., amd64, arm64)
USAGE
}

while getopts "h:m:a:" opt; do
  case "$opt" in
    h) HOST="$OPTARG" ;;
    m) MODE="$OPTARG" ;;
    a) TARGET_ARCH="$OPTARG" ;;
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

echo "[2/6] Build linux/${TARGET_ARCH} binary"
cd "$ROOT_DIR"
mkdir -p build
GOOS=linux GOARCH=${TARGET_ARCH} CGO_ENABLED=0 go build -o build/notion-notifier ./cmd/notion-notifier

# For SSH commands, we use \$HOME so it expands on the remote side.
# For SCP commands, we use ~ which is expanded natively by OpenSSH.
REMOTE_BIN="\$HOME/.local/bin"
REMOTE_CONFIG="\$HOME/.config/notion-notifier"
REMOTE_DATA="\$HOME/.local/share/notion-notifier"
REMOTE_RELEASES="\$HOME/.local/share/notion-notifier/releases"
SCP_RELEASES="~/.local/share/notion-notifier/releases"

echo "[3/6] Prepare remote directories & Migrate old structure"
ssh "$HOST" "
  # Migration from old monolithic structure (~/.local/share/notion-notifier/shared -> XDG dirs)
  if [[ -d \"$REMOTE_DATA/shared\" && ! -f \"$REMOTE_CONFIG/config.yaml\" ]]; then
    echo 'Migrating old directory structure to XDG...'
    mkdir -p $REMOTE_CONFIG
    [ -f \"$REMOTE_DATA/shared/config.yaml\" ] && mv \"$REMOTE_DATA/shared/config.yaml\" \"$REMOTE_CONFIG/\"
    [ -f \"$REMOTE_DATA/shared/env.yaml\" ] && mv \"$REMOTE_DATA/shared/env.yaml\" \"$REMOTE_CONFIG/\"
    [ -f \"$REMOTE_DATA/shared/data.db\" ] && mv \"$REMOTE_DATA/shared/data.db\" \"$REMOTE_DATA/\"
    # Safely move any remaining custom user files (like TLS custom certificates) to data dir before deletion
    mv \"$REMOTE_DATA/shared/\"* \"$REMOTE_DATA/\" 2>/dev/null || true
    rm -rf \"$REMOTE_DATA/shared\"
  fi

  # Create standard directories
  mkdir -p $REMOTE_BIN
  mkdir -p $REMOTE_CONFIG
  mkdir -p $REMOTE_DATA
  mkdir -p $REMOTE_RELEASES/$RELEASE_TS
  mkdir -p ~/.config/systemd/user
"

echo "[4/6] Upload artifacts and runtime files"
scp "$ROOT_DIR/build/notion-notifier" "$HOST:$SCP_RELEASES/$RELEASE_TS/notion-notifier"
if [[ "$MODE" == "install" ]]; then
  # Only upload config/env/db if they don't already exist to prevent overwriting during 'install' mode
  ssh "$HOST" "
    [ ! -f \"$REMOTE_CONFIG/config.yaml\" ] && echo 'Uploading config.yaml'
  "
  # We use a temporary script approach to safely upload without overwrite.
  scp "$ROOT_DIR/config.yaml" "$HOST:$SCP_RELEASES/$RELEASE_TS/config.yaml.default"
  scp "$ROOT_DIR/env.yaml" "$HOST:$SCP_RELEASES/$RELEASE_TS/env.yaml.default"
  scp "$ROOT_DIR/data.db" "$HOST:$SCP_RELEASES/$RELEASE_TS/data.db.default"
fi

echo "[5/6] Set permissions and switch release"
if [[ "$MODE" == "install" ]]; then
  ssh "$HOST" "
  # Copy defaults if physical files don't exist
  [ ! -f \"$REMOTE_CONFIG/config.yaml\" ] && cp \"$REMOTE_RELEASES/$RELEASE_TS/config.yaml.default\" \"$REMOTE_CONFIG/config.yaml\"
  [ ! -f \"$REMOTE_CONFIG/env.yaml\" ] && cp \"$REMOTE_RELEASES/$RELEASE_TS/env.yaml.default\" \"$REMOTE_CONFIG/env.yaml\"
  [ ! -f \"$REMOTE_DATA/data.db\" ] && cp \"$REMOTE_RELEASES/$RELEASE_TS/data.db.default\" \"$REMOTE_DATA/data.db\"
  
  chmod 700 $REMOTE_CONFIG $REMOTE_DATA
  chmod 600 $REMOTE_CONFIG/config.yaml $REMOTE_CONFIG/env.yaml $REMOTE_DATA/data.db 2>/dev/null || true
  chmod 755 $REMOTE_RELEASES/$RELEASE_TS $REMOTE_RELEASES/$RELEASE_TS/notion-notifier
  ln -sfn $REMOTE_RELEASES/$RELEASE_TS/notion-notifier $REMOTE_BIN/notion-notifier
  "
else
  ssh "$HOST" "
  chmod 755 $REMOTE_RELEASES/$RELEASE_TS $REMOTE_RELEASES/$RELEASE_TS/notion-notifier
  ln -sfn $REMOTE_RELEASES/$RELEASE_TS/notion-notifier $REMOTE_BIN/notion-notifier
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
WorkingDirectory=%h/.local/share/notion-notifier
ExecStart=%h/.local/bin/notion-notifier -config %h/.config/notion-notifier/config.yaml -env %h/.config/notion-notifier/env.yaml -db %h/.local/share/notion-notifier/data.db
Restart=always
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ReadWritePaths=%h/.local/share/notion-notifier %h/.config/notion-notifier

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
