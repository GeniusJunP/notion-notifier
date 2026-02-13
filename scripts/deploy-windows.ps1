param(
  [string]$HostName = "example-server"
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$ReleaseTs = Get-Date -Format "yyyyMMddHHmmss"
$RemoteBase = "~/.local/share/notion-notifier"

Write-Host "[1/6] Build web assets"
Push-Location "$Root\web"
npm ci
npm run build
Pop-Location

Write-Host "[2/6] Build linux/arm64 binary"
Push-Location $Root
if (-not (Test-Path "$Root\build")) {
  New-Item -ItemType Directory -Path "$Root\build" | Out-Null
}
$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"
go build -o build/notion-notifier ./cmd/notion-notifier
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
Remove-Item Env:CGO_ENABLED
Pop-Location

Write-Host "[3/6] Prepare remote directories"
ssh $HostName "mkdir -p $RemoteBase/releases/$ReleaseTs $RemoteBase/shared ~/.config/systemd/user"

Write-Host "[4/6] Upload artifacts and runtime files"
scp "$Root/build/notion-notifier" "$HostName`:$RemoteBase/releases/$ReleaseTs/notion-notifier"
scp "$Root/config.yaml" "$HostName`:$RemoteBase/shared/config.yaml"
scp "$Root/env.yaml" "$HostName`:$RemoteBase/shared/env.yaml"
scp "$Root/data.db" "$HostName`:$RemoteBase/shared/data.db"

Write-Host "[5/6] Set permissions and switch release"
ssh $HostName "chmod 700 $RemoteBase $RemoteBase/shared && chmod 755 $RemoteBase/releases/$ReleaseTs $RemoteBase/releases/$ReleaseTs/notion-notifier && chmod 600 $RemoteBase/shared/config.yaml $RemoteBase/shared/env.yaml $RemoteBase/shared/data.db && ln -sfn $RemoteBase/releases/$ReleaseTs $RemoteBase/current"

Write-Host "[6/6] Install/restart user service"
$serviceScript = @'
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
'@

ssh $HostName $serviceScript

Write-Host "Deployment completed: $ReleaseTs"
