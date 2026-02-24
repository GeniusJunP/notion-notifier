#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# If the user didn't specify '-a', we want to default to amd64 when deploying from Linux.
# To do this cleanly, we scan the arguments. If '-a' is not found, we append '-a amd64'.
HAS_ARCH=0
for arg in "$@"; do
  if [[ "$arg" == "-a" ]]; then
    HAS_ARCH=1
    break
  fi
done

if [[ "$HAS_ARCH" -eq 0 ]]; then
  exec "$SCRIPT_DIR/deploy-mac.sh" "$@" -a amd64
else
  exec "$SCRIPT_DIR/deploy-mac.sh" "$@"
fi
