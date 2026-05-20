param(
    [switch]$SkipLaunch,
    [switch]$NoDeps
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$AppName = "weazlfeed"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$InstallRoot = if ($env:WEAZLFEED_HOME) { $env:WEAZLFEED_HOME } else { Join-Path $env:APPDATA "WeazlFeed" }
$BinDir = Join-Path $InstallRoot "bin"
$ConfigDir = Join-Path $InstallRoot "config"
$VaultDir = Join-Path $InstallRoot "vaults"
$CacheDir = Join-Path $InstallRoot "cache"
$GoCache = if ($env:GOCACHE) { $env:GOCACHE } else { Join-Path $CacheDir "go-build" }
$GoModCache = if ($env:GOMODCACHE) { $env:GOMODCACHE } else { Join-Path $CacheDir "go-mod" }
$MsysRoot = "C:\msys64"
$MsysBin = Join-Path $MsysRoot "usr\bin"
$UcrtBin = Join-Path $MsysRoot "ucrt64\bin"

$Binaries = @(
    @{ Name = "weazlfeed"; Cmd = ".\cmd\weazlfeed" },
    @{ Name = "weazlfeed-setup"; Cmd = ".\cmd\weazlfeed-setup" },
    @{ Name = "weazlfeed-import"; Cmd = ".\cmd\weazlfeed-import" },
    @{ Name = "weazlfeed-refresh"; Cmd = ".\cmd\weazlfeed-refresh" },
    @{ Name = "weazlfeed-podcast-search"; Cmd = ".\cmd\weazlfeed-podcast-search" },
    @{ Name = "weazlfeed-prune"; Cmd = ".\cmd\weazlfeed-prune" },
    @{ Name = "weazlfeed-vault"; Cmd = ".\cmd\weazlfeed-vault" }
)

function Test-Command($Name) {
    return $null -ne (Get-Command $Name -ErrorAction SilentlyContinue)
}

function Add-PathEntry($Path) {
    if (!(Test-Path $Path)) {
        return
    }
    $entries = $env:PATH -split ";" | Where-Object { $_ }
    if ($entries -notcontains $Path) {
        $env:PATH = "$Path;$env:PATH"
    }
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $userEntries = @()
    if ($userPath) {
        $userEntries = $userPath -split ";" | Where-Object { $_ }
    }
    if ($userEntries -notcontains $Path) {
        $next = if ($userPath) { "$Path;$userPath" } else { $Path }
        [Environment]::SetEnvironmentVariable("Path", $next, "User")
        Write-Host "Added $Path to the user PATH"
    }
}

function Install-WingetPackage($Id, $Label) {
    if (!(Test-Command "winget")) {
        throw "winget is required to install $Label. Install App Installer from the Microsoft Store, then rerun this script."
    }
    Write-Host "Installing $Label with winget..."
    winget install --id $Id --exact --silent --accept-package-agreements --accept-source-agreements
}

function Ensure-MsysGcc {
    if (Test-Command "gcc") {
        return
    }
    if (!(Test-Path (Join-Path $MsysBin "bash.exe"))) {
        Install-WingetPackage "MSYS2.MSYS2" "MSYS2"
    }
    Add-PathEntry $UcrtBin
    if (!(Test-Path (Join-Path $UcrtBin "gcc.exe"))) {
        $Bash = Join-Path $MsysBin "bash.exe"
        if (!(Test-Path $Bash)) {
            throw "MSYS2 installed, but $Bash was not found. Open a new PowerShell and rerun this script."
        }
        Write-Host "Installing UCRT64 GCC for go-sqlite3..."
        & $Bash -lc "pacman -Sy --needed --noconfirm mingw-w64-ucrt-x86_64-gcc"
    }
    Add-PathEntry $UcrtBin
}

function Ensure-Deps {
    if ($NoDeps) {
        return
    }
    if (!(Test-Command "git")) {
        Install-WingetPackage "Git.Git" "Git"
        Add-PathEntry "C:\Program Files\Git\cmd"
    }
    if (!(Test-Command "go")) {
        Install-WingetPackage "GoLang.Go" "Go"
        Add-PathEntry "C:\Program Files\Go\bin"
    }
    Ensure-MsysGcc
    if (!(Test-Command "ffmpeg")) {
        Install-WingetPackage "Gyan.FFmpeg" "FFmpeg"
    }
    if (!(Test-Command "mpv")) {
        Install-WingetPackage "shinchiro.mpv" "mpv"
    }
}

function Assert-Go {
    if (!(Test-Command "go")) {
        throw "Go is not on PATH. Open a new PowerShell or install Go, then rerun this script."
    }
    if (!(Test-Command "gcc")) {
        throw "gcc is not on PATH. go-sqlite3 needs CGO. Rerun this script without -NoDeps or install MSYS2 UCRT64 GCC."
    }
}

New-Item -ItemType Directory -Force -Path $BinDir, $ConfigDir, $VaultDir, $GoCache, $GoModCache | Out-Null
Ensure-Deps
Add-PathEntry $BinDir
Assert-Go

Write-Host "Building WeazlFeed..."
Push-Location $RepoRoot
try {
    foreach ($Binary in $Binaries) {
        $Out = Join-Path $BinDir ($Binary.Name + ".exe")
        $env:GOCACHE = $GoCache
        $env:GOMODCACHE = $GoModCache
        $env:CGO_ENABLED = "1"
        go build -buildvcs=false -o $Out $Binary.Cmd
        Write-Host "Installed $($Binary.Name) to $Out"
    }
}
finally {
    Pop-Location
}

Write-Host ""
Write-Host "WeazlFeed Windows paths:"
Write-Host "  bin:    $BinDir"
Write-Host "  config: $ConfigDir"
Write-Host "  vaults: $VaultDir"
Write-Host ""
Write-Host "Configuring local model provider..."
& (Join-Path $BinDir "weazlfeed-setup.exe")

if ($SkipLaunch -or $env:WEAZLFEED_SKIP_LAUNCH -eq "1") {
    Write-Host "Skipping first launch."
} else {
    Write-Host "Launching WeazlFeed..."
    & (Join-Path $BinDir "weazlfeed.exe")
}
