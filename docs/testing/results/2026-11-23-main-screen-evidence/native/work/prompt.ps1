param([string]$label)
$identityRoot=Split-Path -Parent $PSCommandPath
function global:prompt { '> ' }
Set-PSReadLineOption -HistorySavePath (Join-Path $identityRoot ($label.Replace(' ','_')+'.history.txt'))
$p=Get-Process -Id $PID
@{label=$label;pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path;cwd=$PWD.Path}|ConvertTo-Json|Set-Content (Join-Path $identityRoot ($label.Replace(' ','_')+'.identity.json')) -Encoding utf8
Clear-Host
'SCROLL_NEEDLE_中文_α'
1..40 | ForEach-Object { 'history line '+$_ }
Write-Output ('COPY_READY '+$label)
