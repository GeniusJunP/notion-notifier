<#
.SYNOPSIS
Installs Notion Notifier on Windows from GitHub Releases.

.DESCRIPTION
This script fetches the latest Windows executable from GitHub Releases,
sets up the application data directories in %LOCALAPPDATA%,
and creates a background startup task using Windows Task Scheduler so it
runs automatically when the user logs in.
#>

$ErrorActionPreference = "Stop"

$Repo = "YOUR_GITHUB_USERNAME/notion-notifier" # TODO: リポジトリの所有者に変更してください
$BinName = "notion-notifier"

Write-Host "=> Installing $BinName from GitHub Releases ($Repo)..." -ForegroundColor Cyan

# 1. アーキテクチャの判定
$Arch = $Env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "AMD64") {
    $DlArch = "amd64"
} elseif ($Arch -eq "ARM64") {
    $DlArch = "arm64"
} else {
    Write-Host "Error: Unsupported architecture $Arch" -ForegroundColor Red
    Exit 1
}

$AssetName = "${BinName}-windows-${DlArch}.exe"

# 2. 最新リリースのダウンロードURLを取得
Write-Host "=> Fetching latest release information..."
$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"

try {
    $ReleaseData = Invoke-RestMethod -Uri $ApiUrl -UseBasicParsing
    $DownloadUrl = ($ReleaseData.assets | Where-Object name -eq $AssetName).browser_download_url
} catch {
    Write-Host "Failed to fetch release data. Are you rate-limited?" -ForegroundColor Yellow
    $DownloadUrl = $null
}

if (-not $DownloadUrl) {
    Write-Host "Error: Could not find release asset '$AssetName'." -ForegroundColor Red
    Write-Host "1. Please ensure $Repo is a public repository."
    Write-Host "2. Please ensure a release exists with the attached binary."
    Exit 1
}

# 3. ディレクトリの作成
$BaseDir = Join-Path $Env:LOCALAPPDATA $BinName
$BinDir = Join-Path $BaseDir "bin"
$ConfigDir = Join-Path $BaseDir "config"
$DataDir = Join-Path $BaseDir "data"
$LogDir = Join-Path $BaseDir "logs"

Write-Host "=> Creating directories in $BaseDir..."
$null = New-Item -ItemType Directory -Force -Path $BinDir, $ConfigDir, $DataDir, $LogDir

# 4. バイナリのダウンロード
$ExePath = Join-Path $BinDir "${BinName}.exe"
Write-Host "=> Downloading $AssetName..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $ExePath

# 5. 設定ファイル・DBの初期化 (既に存在する場合はスキップ)
$ConfigYaml = Join-Path $ConfigDir "config.yaml"
$EnvYaml = Join-Path $ConfigDir "env.yaml"
$DbFile = Join-Path $DataDir "data.db"

if (-not (Test-Path $ConfigYaml)) {
    Write-Host "=> Creating dummy config.yaml..."
    @"
server:
  port: 18080
auth:
  username: "admin"
  password: "password"
"@ | Out-File -FilePath $ConfigYaml -Encoding UTF8
}

if (-not (Test-Path $EnvYaml)) {
    Write-Host "=> Creating dummy env.yaml..."
    @"
notion_api_key: ""
database_id: ""
google_credentials_json: ""
"@ | Out-File -FilePath $EnvYaml -Encoding UTF8
}

if (-not (Test-Path $DbFile)) {
    Write-Host "=> Initializing empty data.db..."
    $null = New-Item -ItemType File -Force -Path $DbFile
}

# 6. タスクスケジューラへの登録 (Windows起動時にバックグラウンドで実行)
Write-Host "=> Setting up Windows Scheduled Task for startup..."

$TaskName = "NotionNotifierService"
$Action = New-ScheduledTaskAction -Execute $ExePath -Argument "-config `"$ConfigYaml`" -env `"$EnvYaml`" -db `"$DbFile`"" -WorkingDirectory $BaseDir
$Trigger = New-ScheduledTaskTrigger -AtLogOn
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable -DontStopOnIdleEnd -ExecutionTimeLimit 0

# バックグラウンド実行のログ出力をハンドルするための一時的なバッチまたはVBSを使う手もあるが、
# Windows固有の Go コマンドラインアプリはコンソールウィンドウが出るため、
# Start-Process -WindowStyle Hidden 的なラッパーを用意する
$VbsWrapper = Join-Path $BinDir "run-hidden.vbs"
@"
Set objShell = CreateObject("WScript.Shell")
args = "-config ""$ConfigYaml"" -env ""$EnvYaml"" -db ""$DbFile"""
objShell.Run """$ExePath"" " & args, 0, False
"@ | Out-File -FilePath $VbsWrapper -Encoding Ascii

$ActionHidden = New-ScheduledTaskAction -Execute "wscript.exe" -Argument "`"$VbsWrapper`"" -WorkingDirectory $BaseDir

# タスクを登録（既存があれば上書き）
Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue | Out-Null
Register-ScheduledTask -TaskName $TaskName -Action $ActionHidden -Trigger $Trigger -Settings $Settings | Out-Null

Write-Host "=> Starting service now..."
Start-ScheduledTask -TaskName $TaskName

Write-Host ""
Write-Host "===========================================================" -ForegroundColor Green
Write-Host "✅ Installation complete!" -ForegroundColor Green
Write-Host "The service has been started in the background (hidden)."
Write-Host "It will automatically start when you log into Windows."
Write-Host "Configuration files are located at:"
Write-Host "  $ConfigDir"
Write-Host "===========================================================" -ForegroundColor Green
