param([string]$label)
function global:prompt { '> ' }
$p=Get-Process -Id $PID
@{label=$label;pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content ($label.Replace(' ','_')+'.identity.json') -Encoding utf8
Clear-Host
'SCROLL_NEEDLE_中文_α'
1..40 | ForEach-Object { 'history line '+$_ }
Write-Output ('COPY_READY '+$label)
