$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-allmotion-desktop-20261123';$dir=Join-Path $root 'desktop-continuation'
$last=Get-ChildItem (Join-Path $dir 'responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1;if($last.Name -cne '149.json' -or @(Get-ChildItem (Join-Path $dir 'commands') -File).Count){throw 'Controller is not at observed idle boundary'}
$id=([IO.File]::ReadAllText((Join-Path $root 'flow-resume.json'))|ConvertFrom-Json).original_driver;$p=Get-Process -Id $id.pid
if($p.Path -ne $id.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.SessionId -ne 2){throw 'Original controller changed'}
Stop-Process -Id $p.Id -ErrorAction Stop
$until=[DateTime]::UtcNow.AddSeconds(5);while((Get-ScheduledTask 'AsdMainAllMotionDesktop20261123').State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if((Get-ScheduledTask 'AsdMainAllMotionDesktop20261123').State -eq 'Running'){throw 'Controller task still running'}
@{utc=[DateTime]::UtcNow.ToString('o');stopped_controller=$id;last_sequence=149;reason='Replace idle SetCursorPos controller with native SendInput movement; no Run/UI/window restarted';scope='Only exact controller process stopped, no mouse button held and no queued command'}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'controller-replacement.json') -Encoding utf8
Get-Content (Join-Path $root 'controller-replacement.json') -Raw -Encoding utf8
