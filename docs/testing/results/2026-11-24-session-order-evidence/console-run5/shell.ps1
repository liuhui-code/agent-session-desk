param([string]$label)
function global:prompt { '> ' }
Set-PSReadLineOption -HistorySavePath (Join-Path $PSScriptRoot ($label+'.history.txt'))
$p=Get-Process -Id $PID
$identity=Join-Path $PSScriptRoot ($label+'.identity.json')
$json=@{label=$label;pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json
[IO.File]::WriteAllText(($identity+'.pending'),$json,[Text.UTF8Encoding]::new($false));[IO.File]::Move(($identity+'.pending'),$identity)
Clear-Host
'ORDER_READY '+$label
