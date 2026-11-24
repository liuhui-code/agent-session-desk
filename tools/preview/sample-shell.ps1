param([string]$Label)
$ErrorActionPreference = 'Stop'
$data = Join-Path $PSScriptRoot 'data'
Set-PSReadLineOption -HistorySavePath (Join-Path $data ('shell-' + $Label + '.history.txt'))
$process = Get-Process -Id $PID
@{ pid=$PID; filetime=$process.StartTime.ToUniversalTime().ToFileTimeUtc(); label=$Label } |
    ConvertTo-Json | Set-Content -LiteralPath (Join-Path $data ('shell-' + $Label + '.identity.json')) -Encoding utf8
Clear-Host
Write-Host ('PowerShell ' + $Label + ' · 真实交互终端') -ForegroundColor Cyan
Write-Host '可以直接输入命令。F6 切换管理模式，? 查看快捷键。'
Write-Host ''
