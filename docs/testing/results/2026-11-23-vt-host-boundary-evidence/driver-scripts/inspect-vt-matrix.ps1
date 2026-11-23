$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$cases=@()
foreach($case in @('candidate-system1','candidate-system2','candidate-sdk','official-system')){$root=Join-Path $base ('vt-boundary-'+$case);$result=Join-Path $root 'result.json';$cases+=,@{case=$case;exists=(Test-Path $root);result=$(if(Test-Path $result){[IO.File]::ReadAllText($result)|ConvertFrom-Json}else{$null})}}
$paths=@((Join-Path $base 'history-backend-candidate/psmux.exe'),(Join-Path $base 'backend/psmux.exe'))
@{utc=[DateTime]::UtcNow.ToString('o');cases=$cases;live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths}|ForEach-Object{@{pid=$_.ProcessId;path=$_.ExecutablePath}})}|ConvertTo-Json -Depth 12
