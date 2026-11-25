param([string]$Label)
$ErrorActionPreference = 'Stop'
$data = Join-Path $PSScriptRoot 'data'
Set-PSReadLineOption -HistorySavePath (Join-Path $data ('shell-' + $Label + '.history.txt'))
$process = Get-Process -Id $PID
@{ pid=$PID; filetime=$process.StartTime.ToUniversalTime().ToFileTimeUtc(); label=$Label } |
    ConvertTo-Json | Set-Content -LiteralPath (Join-Path $data ('shell-' + $Label + '.identity.json')) -Encoding utf8
Clear-Host
