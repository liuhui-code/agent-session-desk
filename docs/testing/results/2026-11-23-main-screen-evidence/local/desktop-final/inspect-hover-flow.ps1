$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-allmotion-desktop-20261123';$dir=Join-Path $root 'desktop-continuation'
$f=Get-ChildItem (Join-Path $dir 'responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1
$r=[IO.File]::ReadAllText($f.FullName)|ConvertFrom-Json;$ready=[IO.File]::ReadAllText((Join-Path $dir 'ready.json'))|ConvertFrom-Json;$p=Get-Process -Id $ready.pid
if($p.Path -ne $ready.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Original window changed'}
$failure=[IO.File]::ReadAllText((Join-Path $root 'resume-failure.json'))|ConvertFrom-Json
@{utc=[DateTime]::UtcNow.ToString('o');task_state=[string](Get-ScheduledTask 'AsdMainAllMotionDesktop20261123').State;last_response=$f.Name;last_command=$r.command;last_ok=$r.ok;last_state=$r.state.text;last_divider=(@(Get-ChildItem (Join-Path $dir 'responses') -Filter '*.json'|Sort-Object Name|ForEach-Object{$v=[IO.File]::ReadAllText($_.FullName)|ConvertFrom-Json;if($v.command.kind -eq 'divider'){$v}})|Select-Object -Last 1);failure=$failure;window_original_live=$true;ui_processes=([IO.File]::ReadAllText((Join-Path $root 'ui-identities.json'))|ConvertFrom-Json)}|ConvertTo-Json -Depth 10
