<#
.SYNOPSIS
    Lance go vet + go test sur tout le projet (avec cgo et gcc configurés).

.EXAMPLE
    ./scripts/test.ps1
#>
[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

# Localise gcc (comme build.ps1) : Fyne dépend de cgo même pour compiler les tests.
$gcc = (Get-Command gcc -ErrorAction SilentlyContinue).Source
if (-not $gcc) {
    $pkgRoot = Join-Path $env:LOCALAPPDATA "Microsoft\WinGet\Packages"
    if (Test-Path $pkgRoot) {
        $hit = Get-ChildItem $pkgRoot -Recurse -Filter gcc.exe -ErrorAction SilentlyContinue |
            Select-Object -First 1
        if ($hit) { $env:Path = "$(Split-Path $hit.FullName);$env:Path" }
    }
}
$env:CGO_ENABLED = "1"

Write-Host "go vet ./..." -ForegroundColor Cyan
go vet ./...
if ($LASTEXITCODE -ne 0) { Write-Error "go vet a échoué."; exit 1 }

Write-Host "go test ./..." -ForegroundColor Cyan
go test ./...
exit $LASTEXITCODE
