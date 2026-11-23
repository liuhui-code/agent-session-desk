$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'group-rename-desktop-20261123'
$product=Join-Path $base 'asd-group-rename.exe';$bundle=Join-Path $base 'terminal-bundle-20261123/bundle'
$g=@('--terminal-bundle',$bundle,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-group-rename','--state-dir',(Join-Path $root 'state'))
$rows=@((& $product @g list --json|Out-String)|ConvertFrom-Json);if($LASTEXITCODE -ne 0){throw 'Discovery failed'}
$owned=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json
foreach($s in $rows){if(@($owned|Where-Object{$_.session_id -ceq $s.session_id -and $_.run_id -ceq $s.run_id -and $_.backend_id -ceq $s.backend_id}).Count -ne 1){throw 'Unowned target'}}
$task=Get-ScheduledTask -TaskName 'AsdGroupRenameDesktop20261123'
if($task.State -eq 'Running'){throw 'Driver still running'}
@{result='observed';utc=[DateTime]::UtcNow.ToString('o');driver_state=[string]$task.State;live_sessions=$rows;actual_live_count=$rows.Count;first_cleanup_expected_count=3;first_cleanup_failure='Unexpected live set; retained three-session guard in reused harness';first_cleanup_before_any_stop=$true;correction='Require exactly two original owned sessions'}|ConvertTo-Json -Depth 7|Set-Content (Join-Path $root 'cleanup-first-failure.json') -Encoding utf8
@{result='observed';live_count=$rows.Count;driver_state=[string]$task.State}|ConvertTo-Json
