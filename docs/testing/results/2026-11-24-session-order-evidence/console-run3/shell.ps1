param([string]$label)
function global:prompt { '> ' }
Set-PSReadLineOption -HistorySavePath (Join-Path $PSScriptRoot ($label+'.history.txt'))
$p=Get-Process -Id $PID
@{label=$label;pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content (Join-Path $PSScriptRoot ($label+'.identity.json')) -Encoding utf8
Clear-Host
'ORDER_READY '+$label
