[CmdletBinding()]
param([string]$ConptyDirectory)
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'windows-resource-process.ps1')
. (Join-Path $PSScriptRoot 'windows-resource-driver.ps1')
$experiment = [IO.Path]::GetTempPath()
$candidate = Join-Path $experiment 'synthetic conpty host'
$selection = [pscustomobject]@{ mode = 'sdk-candidate'; directory = $candidate }
$plain = [pscustomobject]@{ mode = 'system'; directory = $null }
$shell = (Get-Process -Id $PID).Path
function Assert-ChildHost {
    param([hashtable]$Environment, [string]$Expected, [string]$Label)
    $child = Start-RecordedProcess $shell @('-NoLogo', '-NoProfile', '-Command', '[Console]::WriteLine(($env:PSMUX_CONPTY_DIR ?? "<unset>"))') $Environment
    try {
        if (-not $child.process.WaitForExit(10000)) { $child.process.Kill(); throw 'Environment child timed out.' }
        $output = $child.stdout.GetAwaiter().GetResult().Trim()
        $errorText = $child.stderr.GetAwaiter().GetResult()
        if ($child.process.ExitCode -ne 0 -or $errorText -or $output -cne $Expected) { throw "${Label}: child host environment differs: $output" }
    } finally { $child.process.Dispose() }
}
$old = $env:PSMUX_CONPTY_DIR
try {
    $env:PSMUX_CONPTY_DIR = 'synthetic inherited host'
    Assert-ChildHost (Get-AsdResourceDriverEnvironment 'driver data' 'driver config' $plain) '<unset>' 'Default driver'
    Assert-ChildHost (Get-AsdResourceDriverEnvironment 'driver data' 'driver config' $selection) $candidate 'Candidate driver'
    Assert-ChildHost (Get-MuxEnvironment 'backend data' 'backend config') '<unset>' 'Product backend control'
    'Separate child processes confirmed default/inherited isolation and driver-only host selection. No ConPTY DLL was loaded by this test.'
} finally { $env:PSMUX_CONPTY_DIR = $old }

if ($ConptyDirectory) {
    $pin = (Get-Content -Raw (Join-Path $PSScriptRoot 'windows-dependencies.json') | ConvertFrom-Json).resource_driver_conpty
    $selected = Resolve-AsdResourceDriverHost $ConptyDirectory $pin
    if ($selected.mode -ne 'sdk-candidate') { throw 'Explicit candidate was not selected.' }
    $copy = Join-Path ([IO.Path]::GetTempPath()) ('asd-host-check-' + [Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory $copy | Out-Null
    try {
        foreach ($entry in $pin.files.PSObject.Properties) { Copy-Item (Join-Path $ConptyDirectory $entry.Name) (Join-Path $copy $entry.Name) }
        $file = [IO.File]::OpenWrite((Join-Path $copy 'conpty.dll'))
        try { $file.WriteByte(0); $file.Flush() } finally { $file.Dispose() }
        $rejected = $false
        try { $null = Resolve-AsdResourceDriverHost $copy $pin } catch { $rejected = $true }
        if (-not $rejected) { throw 'Changed host file was accepted.' }
        'Pinned Microsoft host files accepted; mutated private copy rejected before any DLL load.'
    } finally { Remove-Item -LiteralPath $copy -Recurse -Force }
}
