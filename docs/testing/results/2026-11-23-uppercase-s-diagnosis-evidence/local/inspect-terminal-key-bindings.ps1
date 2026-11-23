$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'sidebar-key-diagnosis-20261123'
$task=Get-ScheduledTask 'AsdSidebarKeyDesktop20261123'
$settings=Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json';$c=[IO.File]::ReadAllText($settings)|ConvertFrom-Json
@{task_state=[string]$task.State;terminal_actions=$c.actions;terminal_keybindings=$c.keybindings;legacy_bindings=$c.keyBindings;settings_sha256=(Get-FileHash $settings).Hash}|ConvertTo-Json -Depth 12
