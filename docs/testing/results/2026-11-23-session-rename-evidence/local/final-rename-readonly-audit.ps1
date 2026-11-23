$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$product=Join-Path $base 'asd-rename-surface.exe';$root=Join-Path $base 'session-rename-desktop-20261123'
$g=@('--terminal-bundle',(Join-Path $base 'terminal-bundle-20261123/bundle'),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-session-rename','--state-dir',(Join-Path $root 'state'))
$history=(& $product @g history|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $history.Count -ne 3){throw 'Offline history failed'}
$initial=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json
foreach($r in $history){if(@($initial|Where-Object{$_.session_id -eq $r.session.session_id -and $_.run_id -eq $r.session.run_id}).Count -ne 1 -or !$r.stop_confirmed_at -or $r.session.terminal_state -ne 'unknown'){throw 'Offline history changed identity or invented live state'}}
$b=@($initial|Where-Object{$_.title -ceq 'Copy B'})[0];$named=@($history|Where-Object{$_.session.session_id -eq $b.session_id})[0]
if($named.session.title -cne 'Copy A' -or $named.session.title_revision -ne 3){throw 'Offline history lost current renamed label'}
$paths=@('asd-rename.exe','asd-rename-final.exe','asd-rename-surface.exe','rename-session-tests.exe','rename-store-tests.exe','rename-manager-tests.exe','rename-ui-tests.exe','rename-ui-tests-final.exe','rename-ui-tests-surface.exe','rename-ui-tests-complete.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count){throw 'Private image remains'}
$tasks=@('AsdSessionRenameSetup20261123','AsdSessionRenameDesktop20261123','AsdSessionRenameSurfaceSetup20261123','AsdSessionRenameSurfaceDesktop20261123');foreach($task in $tasks){if(Get-ScheduledTask $task -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
$settings=Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json';$hash=(Get-FileHash $settings).Hash;if($hash -ne '1499307ABEBA68C23F9EB38A4C636F81172475B6B2B535BC0B11DCD9547241F8'){throw 'Settings differs'}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');offline_history=$history;private_images_alive=$live.Count;tasks_absent=$tasks;settings_sha256=$hash;scope='Read-only final audit after all suites and desktop cleanup; no commands replayed, no live backend started'}|ConvertTo-Json -Depth 9
