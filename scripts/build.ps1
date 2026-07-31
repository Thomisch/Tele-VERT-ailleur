<#
.SYNOPSIS
    Compile FuckTeamsStatus en un exécutable Windows autonome (sans console).

.DESCRIPTION
    Fyne nécessite cgo + un compilateur C (gcc). Ce script localise gcc
    (MinGW/winlibs installé via winget si présent), active CGO, puis compile
    cmd/fuckteamsstatus vers bin/FuckTeamsStatus.exe avec le flag -H=windowsgui
    (pas de fenêtre console) et -s -w (binaire allégé).

.EXAMPLE
    ./scripts/build.ps1
#>
[CmdletBinding()]
param(
    [string]$Output = "bin/TeleVertAilleur.exe"
)

$ErrorActionPreference = "Stop"

# Racine du repo = dossier parent de scripts/
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

# --- Localise gcc ---
$gcc = (Get-Command gcc -ErrorAction SilentlyContinue).Source
if (-not $gcc) {
    $pkgRoot = Join-Path $env:LOCALAPPDATA "Microsoft\WinGet\Packages"
    if (Test-Path $pkgRoot) {
        $hit = Get-ChildItem $pkgRoot -Recurse -Filter gcc.exe -ErrorAction SilentlyContinue |
            Select-Object -First 1
        if ($hit) {
            $env:Path = "$(Split-Path $hit.FullName);$env:Path"
            $gcc = $hit.FullName
        }
    }
}
if (-not $gcc) {
    Write-Error "gcc introuvable. Installe MinGW, p.ex. : winget install BrechtSanders.WinLibs.POSIX.UCRT"
    exit 1
}
Write-Host "gcc : $gcc" -ForegroundColor DarkGray

# --- Build ---
$env:CGO_ENABLED = "1"
$outDir = Split-Path -Parent $Output
if ($outDir -and -not (Test-Path $outDir)) { New-Item -ItemType Directory -Force -Path $outDir | Out-Null }

Write-Host "Compilation -> $Output" -ForegroundColor Cyan
go build -ldflags "-H=windowsgui -s -w" -o $Output ./cmd/fuckteamsstatus
if ($LASTEXITCODE -ne 0) { Write-Error "Échec de la compilation."; exit 1 }

$size = [math]::Round((Get-Item $Output).Length / 1MB, 2)
Write-Host "OK : $Output ($size Mo)" -ForegroundColor Green
