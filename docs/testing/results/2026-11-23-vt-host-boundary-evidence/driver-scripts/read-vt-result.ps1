$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$r=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/history-desktop-20261123'
@{task_state=[string](Get-ScheduledTask AsdHistoryDesktop20261123).State;vt=([IO.File]::ReadAllText((Join-Path $r 'work/alt-observed.json'))|ConvertFrom-Json);lifecycle_failure=([IO.File]::ReadAllText((Join-Path $r 'lifecycle-failure.json'))|ConvertFrom-Json)}|ConvertTo-Json -Depth 7
