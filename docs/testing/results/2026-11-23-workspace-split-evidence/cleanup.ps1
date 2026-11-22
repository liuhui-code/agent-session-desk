$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'layout-20261123'
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName}).Count -ne 0){throw 'Private window still running'}
foreach($name in @('AsdLayout20261123','AsdInjectedDrag20261123','AsdInjectedDrag3_20261123')){$task=Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue;if($task -and $task.State -eq 'Running'){throw 'Owned task still running'}}
$product=Join-Path $base 'layout-ui.exe';$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-layout','--state-dir',(Join-Path $root 'state'))
$rowsText=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'List failed'};$rows=$rowsText|ConvertFrom-Json
$owned=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json;$stops=@()
foreach($v in $rows){if(@($owned|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id}).Count -ne 1){throw 'Unexpected target'};$text=(& $product @g stop $v.session_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Stop failed'};$stops+=,@{session_id=$v.session_id;run_id=$v.run_id;result=$text}}
$list=(& $product @g list --json|Out-String).Trim();if($LASTEXITCODE -ne 0 -or $list -ne '[]'){throw 'Owned sessions remain'}
$r=[IO.File]::ReadAllText((Join-Path $root 'receipt.json'))|ConvertFrom-Json
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';if((Get-FileHash $settings).Hash -ne $r.settings_test){throw 'Settings changed externally'}
[IO.File]::Copy((Join-Path $root 'settings.before'),$settings,$true);if((Get-FileHash $settings).Hash -ne $r.settings_before){throw 'Restore mismatch'}
foreach($name in @('AsdLayout20261123','AsdInjectedDrag20261123','AsdInjectedDrag3_20261123')){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){Unregister-ScheduledTask -TaskName $name -Confirm:$false}}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');stops=$stops;remaining_sessions=$list;private_windows=0;settings_restored=(Get-FileHash $settings).Hash}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'cleanup.json') -Encoding utf8
Get-Content (Join-Path $root 'cleanup.json') -Raw -Encoding utf8
