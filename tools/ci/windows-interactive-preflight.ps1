[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)][string]$ProductBinary,
    [Parameter(Mandatory = $true)][string]$PsmuxBinary,
    [Parameter(Mandatory = $true)][string]$PreviewManifest,
    [string]$OutputDirectory = 'build/manual-windows-acceptance'
)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if (-not $IsWindows) { throw 'This preflight requires native Windows.' }
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$lock = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'windows-dependencies.json') -Raw | ConvertFrom-Json
$manifest = Get-Content -LiteralPath $PreviewManifest -Raw | ConvertFrom-Json
$product = (Resolve-Path -LiteralPath $ProductBinary).Path
$psmux = (Resolve-Path -LiteralPath $PsmuxBinary).Path
$productHash = (Get-FileHash -LiteralPath $product -Algorithm SHA256).Hash.ToLowerInvariant()
$psmuxHash = (Get-FileHash -LiteralPath $psmux -Algorithm SHA256).Hash.ToLowerInvariant()
if ($manifest.os -cne 'windows' -or $manifest.arch -cne 'amd64') { throw 'Manifest target is not windows/amd64.' }
if ($manifest.sha256 -cne $productHash) { throw 'Product binary does not match preview manifest.' }
if ($psmuxHash -cne $lock.psmux.executable_sha256) { throw 'psmux does not match pinned official executable.' }
if ($manifest.psmux_sha256_required -cne $psmuxHash) { throw 'Manifest and pinned psmux hash disagree.' }
$versionText = (& $product version 2>&1 | Out-String).Trim()
if ($LASTEXITCODE -ne 0) { throw 'Product version command failed.' }
$backendText = (& $psmux -V 2>&1 | Out-String).Trim()
if ($LASTEXITCODE -ne 0) { throw 'psmux version command failed.' }
$os = Get-CimInstance Win32_OperatingSystem
$terminal = $null
try {
    $terminal = Get-AppxPackage -Name Microsoft.WindowsTerminal -ErrorAction Stop |
        Select-Object -First 1 -ExpandProperty Version
} catch { $terminal = $null }
$out = [IO.Path]::GetFullPath($OutputDirectory)
New-Item -ItemType Directory -Force -Path $out | Out-Null
$record = [ordered]@{
    schema_version = 1
    collected_utc = [DateTime]::UtcNow.ToString('o')
    product_commit = [string]$manifest.commit
    product_sha256 = $productHash
    product_version = $versionText
    psmux_sha256 = $psmuxHash
    psmux_version_output = $backendText
    os_caption = [string]$os.Caption
    os_version = [string]$os.Version
    os_build = [string]$os.BuildNumber
    architecture = [Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
    windows_terminal_package_version = $(if ($null -ne $terminal) { [string]$terminal } else { $null })
    terminal_package_detection_note = 'Package may be absent or unavailable; record actual host/version manually.'
    preflight_result = 'pass'
    manual_tests_result = 'not_run'
    notes = 'No agent credentials, environment variables, commands, process arguments, runtime paths, or user directories collected.'
}
$target = Join-Path $out 'preflight.json'
$record | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $target -Encoding utf8
Write-Host "Preflight passed. Metadata saved to $target"
Write-Host 'Interactive Windows Terminal, IME, attach/detach, focus and resize tests are NOT run.'
