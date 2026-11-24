[CmdletBinding()]
param([switch]$InTerminal)
$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot
try {
    $manifest = Get-Content -Raw -LiteralPath (Join-Path $root 'alpha.json') | ConvertFrom-Json
    foreach ($entry in $manifest.files.PSObject.Properties) {
        $file = Join-Path $root $entry.Name
        if ((Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant() -cne $entry.Value) {
            throw ('Preview file hash mismatch: ' + $entry.Name)
        }
    }
    $hostConfig = Get-Content -Raw -LiteralPath (Join-Path $root 'local-host.json') | ConvertFrom-Json
    foreach ($hostName in @('terminal', 'shell')) {
        $entry = $hostConfig.$hostName
        if ((Get-FileHash -LiteralPath $entry.path -Algorithm SHA256).Hash.ToLowerInvariant() -cne $entry.sha256) {
            throw ('Preview host changed: ' + $hostName)
        }
    }
    if (!$InTerminal) {
        $profile = ''
        if ($hostConfig.profileGuid) { $profile = ' -p "' + $hostConfig.profileGuid + '"' }
        $arguments = '-w new' + $profile + ' --title "Agent Session Desk Alpha" --startingDirectory "' + $root + '" "' +
            $hostConfig.shell.path + '" -NoLogo -NoProfile -File "' + $PSCommandPath + '" -InTerminal'
        Start-Process -FilePath $hostConfig.terminal.path -ArgumentList $arguments | Out-Null
        return
    }
    $productName = 'asd.exe'
    if ($manifest.entrypoint) { $productName = $manifest.entrypoint }
    $product = Join-Path $root $productName
    $data = Join-Path $root 'data'
    $workspace = Join-Path $root 'workspace'
    New-Item -ItemType Directory -Force -Path $data, $workspace | Out-Null
    $globalArguments = @('--terminal-bundle', (Join-Path $root 'terminal-runtime'),
        '--namespace', 'asd-alpha-8234c4e', '--runtime-dir', (Join-Path $data 'runtime'),
        '--state-dir', (Join-Path $data 'state'))
    $text = & $product @globalArguments list --json | Out-String
    if ($LASTEXITCODE -ne 0) { throw 'Could not discover preview sessions; no launch retried.' }
    $sessions = $text | ConvertFrom-Json
    if ($null -eq $sessions -or @($sessions).Count -eq 0) {
        foreach ($label in @('A', 'B')) {
            $result = & $product @globalArguments start --name ('PowerShell ' + $label) --cwd $workspace -- `
                $hostConfig.shell.path -NoLogo -NoProfile -NoExit -File (Join-Path $root 'sample-shell.ps1') $label
            if ($LASTEXITCODE -ne 0) { throw ('Could not start sample ' + $label + '; no launch retried.') }
            $result | Set-Content -LiteralPath (Join-Path $data ('sample-' + $label + '.json')) -Encoding utf8
        }
    }
    & $product @globalArguments
    if ($LASTEXITCODE -ne 0) { throw ('Preview exited with code ' + $LASTEXITCODE) }
} catch {
    Write-Host ('启动失败：' + $_.Exception.Message) -ForegroundColor Red
    Read-Host '按 Enter 关闭；请保留 data 目录以便排查' | Out-Null
    exit 1
}
