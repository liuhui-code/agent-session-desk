$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$out=@{}
foreach($n in @('asd-main-allmotion.exe','main-allmotion-ui.exe','asd-main-mouse-probe.exe')){$out[$n]=(Get-FileHash (Join-Path $base $n)).Hash.ToLowerInvariant()}
$out|ConvertTo-Json
