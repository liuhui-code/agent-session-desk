$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'footer-input-desktop-20261123';$desktop=Join-Path $root 'desktop'
$task=Get-ScheduledTask 'AsdFooterInputDesktop20261123';$argsExact=$task.Actions[0].Arguments
$drivers=@(Get-CimInstance Win32_Process|Where-Object{$_.CommandLine -and $_.CommandLine.Contains($argsExact)})
$live=@();foreach($p in $drivers){$v=Get-Process -Id $p.ProcessId;$live+=,@{pid=$v.Id;filetime=$v.StartTime.ToUniversalTime().ToFileTimeUtc();path=$v.Path;session=$v.SessionId}}
$last=Get-ChildItem (Join-Path $desktop 'responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1
$r=[IO.File]::ReadAllText($last.FullName)|ConvertFrom-Json
$ready=[IO.File]::ReadAllText((Join-Path $desktop 'ready.json'))|ConvertFrom-Json;$window=Get-Process -Id $ready.pid
if($window.Path -ne $ready.path -or $window.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Original window changed'}
@{task_state=[string]$task.State;driver_processes=$live;last_response=$last.Name;response=$r;failure=([IO.File]::ReadAllText((Join-Path $root 'flow-failure.json'))|ConvertFrom-Json);ui_identities=([IO.File]::ReadAllText((Join-Path $root 'ui-identities.json'))|ConvertFrom-Json);window_original_live=$true;utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Depth 15
