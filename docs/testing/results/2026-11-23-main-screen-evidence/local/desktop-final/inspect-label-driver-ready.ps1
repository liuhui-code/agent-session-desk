$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-allmotion-desktop-20261123/desktop'
$task=Get-ScheduledTask 'AsdMainAllMotionDesktop20261123';if($task.State -ne 'Running'){throw 'Driver not running'}
$f=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(15);while(!(Test-Path $f) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $f)){throw 'Ready not observed'}
Get-Content $f -Raw -Encoding utf8
