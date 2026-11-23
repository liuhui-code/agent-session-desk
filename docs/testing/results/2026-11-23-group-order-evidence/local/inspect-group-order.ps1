$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'group-order-desktop-20261123'
$task=Get-ScheduledTask -TaskName 'AsdGroupOrderDesktop20261123'
$failure=[IO.File]::ReadAllText((Join-Path $root 'flow-failure.json'))|ConvertFrom-Json
$last=Get-ChildItem (Join-Path $root 'desktop/responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1
$response=[IO.File]::ReadAllText($last.FullName)|ConvertFrom-Json
$metadata=(& (Join-Path $base 'read-group-metadata.exe') (Join-Path $root 'state/metadata.sqlite3')|Out-String)|ConvertFrom-Json
@{task_state=[string]$task.State;failure=$failure;last_response=$last.Name;command=$response.command;state=$response.state;metadata=$metadata}|ConvertTo-Json -Depth 14
