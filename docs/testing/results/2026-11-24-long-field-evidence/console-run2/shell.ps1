function global:prompt { '> ' }
Set-PSReadLineOption -HistorySavePath (Join-Path $PSScriptRoot 'shell.history.txt')
$p=Get-Process -Id $PID
@{pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content (Join-Path $PSScriptRoot 'shell-identity.json') -Encoding utf8
Clear-Host
'LONG_SHELL_READY'
