[CmdletBinding()]
param(
    [string]$Destination = 'build/ci-tools/psmux-3.4.0'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if (-not $IsWindows) { throw 'This installer prepares the pinned Windows test dependency only.' }

$lock = Get-Content -Raw (Join-Path $PSScriptRoot 'windows-dependencies.json') | ConvertFrom-Json
$destinationPath = [IO.Path]::GetFullPath($Destination)
New-Item -ItemType Directory -Force $destinationPath | Out-Null
$archive = Join-Path $destinationPath 'release.zip'
if (-not (Test-Path -LiteralPath $archive)) {
    Invoke-WebRequest -Uri $lock.psmux.archive_url -OutFile $archive
}
if ((Get-Item -LiteralPath $archive).Length -ne $lock.psmux.archive_size) {
    throw 'psmux archive size does not match the pinned official asset.'
}
if ((Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant() -ne $lock.psmux.archive_sha256) {
    throw 'psmux archive SHA-256 does not match the pinned official asset.'
}
# This directory is a caller-selected test dependency directory, never a global
# installation or a user psmux configuration directory.
$expanded = Join-Path $destinationPath 'release'
Expand-Archive -LiteralPath $archive -DestinationPath $expanded -Force
$binary = Join-Path $expanded $lock.psmux.executable
if ((Get-FileHash -LiteralPath $binary -Algorithm SHA256).Hash.ToLowerInvariant() -ne $lock.psmux.executable_sha256) {
    throw 'Extracted psmux executable SHA-256 does not match the verified release.'
}
Write-Output $binary
