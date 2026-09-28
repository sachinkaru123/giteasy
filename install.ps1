# giteasy installer for Windows.
#
# Usage (PowerShell):
#   irm https://raw.githubusercontent.com/sachinkaru123/giteasy/main/install.ps1 | iex
#
# Downloads the correct prebuilt binary from the latest GitHub Release
# and installs it onto your user PATH. No Go required.

$ErrorActionPreference = "Stop"

# ---- Configure this before publishing your repo -----------------------
$Repo = "sachinkaru123/giteasy"   # <-- change to "your-github-username/giteasy"
# -------------------------------------------------------------------------

$BinaryName = "giteasy.exe"

function Info($msg)    { Write-Host "-> $msg" -ForegroundColor Cyan }
function Ok($msg)      { Write-Host "OK $msg" -ForegroundColor Green }
function Fail($msg)    { Write-Host "FAIL $msg" -ForegroundColor Red; exit 1 }

# ---- Detect architecture --------------------------------------------------
$arch = $env:PROCESSOR_ARCHITECTURE
switch ($arch) {
    "AMD64" { $Arch = "amd64" }
    "ARM64" { $Arch = "arm64" }
    default { Fail "Unsupported architecture: $arch. Try: go install github.com/$Repo@latest" }
}
Info "Detected platform: windows/$Arch"

# ---- Find latest release tag via GitHub API --------------------------------
Info "Looking up latest release ..."
try {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
} catch {
    Fail "Could not reach GitHub API to find the latest release."
}
$Tag = $release.tag_name
if (-not $Tag) {
    Fail "Could not determine the latest release. Check https://github.com/$Repo/releases"
}
Ok "Latest version: $Tag"

# ---- Build expected asset name (matches .goreleaser.yaml naming) ----------
$Asset = "giteasy_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Asset"

$TmpDir = Join-Path $env:TEMP "giteasy-install-$(Get-Random)"
New-Item -ItemType Directory -Path $TmpDir | Out-Null

try {
    Info "Downloading $Asset ..."
    $zipPath = Join-Path $TmpDir $Asset
    Invoke-WebRequest -Uri $Url -OutFile $zipPath -UseBasicParsing

    Info "Extracting ..."
    Expand-Archive -Path $zipPath -DestinationPath $TmpDir -Force

    $exeSource = Join-Path $TmpDir $BinaryName
    if (-not (Test-Path $exeSource)) {
        Fail "Extracted archive did not contain $BinaryName."
    }

    # ---- Install location: per-user, no admin rights required ----------
    $InstallDir = Join-Path $env:LOCALAPPDATA "giteasy"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $exeDest = Join-Path $InstallDir $BinaryName
    Copy-Item -Path $exeSource -Destination $exeDest -Force

    Ok "giteasy $Tag installed to $exeDest"

    # ---- Add to user PATH if missing ------------------------------------
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
        Info "Added $InstallDir to your user PATH."
        Info "Restart your terminal for PATH changes to take effect."
    }

    Write-Host ""
    Ok "Run 'giteasy' in a new terminal to get started."
} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}
