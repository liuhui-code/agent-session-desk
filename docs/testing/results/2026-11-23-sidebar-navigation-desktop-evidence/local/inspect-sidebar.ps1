$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'sidebar-navigation-desktop-20261123'
$task=Get-ScheduledTask 'AsdSidebarNavigationDesktop20261123'
$processes=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq (Join-Path $base 'asd-sidebar-navigation.exe') -or $_.ExecutablePath -eq (Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe')}|Select-Object ProcessId,ExecutablePath,CreationDate)
$last=Get-ChildItem (Join-Path $root 'desktop/responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1
$fail=[IO.File]::ReadAllText((Join-Path $root 'flow-failure.json'))|ConvertFrom-Json
$state=[IO.File]::ReadAllText($last.FullName)|ConvertFrom-Json
@{task_state=[string]$task.State;processes=$processes;last_response=$last.Name;failure=$fail;last_state=$state}|ConvertTo-Json -Depth 10
