$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'visual-compare-20261123'
$task=Get-ScheduledTask -TaskName AsdVisualCompare20261123
$r=@{utc=[DateTime]::UtcNow.ToString('o');task_state=[string]$task.State;frame_receipt=([IO.File]::ReadAllText((Join-Path $root 'frame-receipt.json'))|ConvertFrom-Json);failure=([IO.File]::ReadAllText((Join-Path $root 'flow-failure.json'))|ConvertFrom-Json);fixtures=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq (Join-Path $base 'visual-fixture.exe')}|ForEach-Object{@{pid=$_.ProcessId;creation=$_.CreationDate;path=$_.ExecutablePath}})}
$r|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'live-dimension-check.json') -Encoding utf8
$r|ConvertTo-Json -Depth 6
if($task.State -ne 'Running'){throw 'Original desktop task is not live'}
