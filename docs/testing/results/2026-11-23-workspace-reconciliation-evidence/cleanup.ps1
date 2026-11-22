$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'workspace-reconcile-20261123'
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
$active=@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName})
if($active.Count -ne 0){throw 'Private terminal still running'}
$r=[IO.File]::ReadAllText((Join-Path $root 'receipt.json'))|ConvertFrom-Json
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json'
if((Get-FileHash $settings).Hash -ne $r.settings_test){throw 'Settings changed externally'}
[IO.File]::Copy((Join-Path $root 'settings.before'),$settings,$true)
if((Get-FileHash $settings).Hash -ne $r.settings_before){throw 'Restore mismatch'}
$task=Get-ScheduledTask -TaskName AsdWorkspaceReconcile20261123
if($task.State -eq 'Running'){throw 'Owned driver still running'}
Unregister-ScheduledTask -TaskName AsdWorkspaceReconcile20261123 -Confirm:$false
$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-ui-reconcile','--state-dir',(Join-Path $root 'state'))
$list=(& (Join-Path $base 'reconcile-ui.exe') @g list --json|Out-String).Trim()
if($LASTEXITCODE -ne 0 -or $list -ne '[]'){throw 'Unexpected terminal Run'}
@{utc=[DateTime]::UtcNow.ToString('o');result='passed';private_windows=0;remaining_sessions=$list;settings_restored=(Get-FileHash $settings).Hash}|ConvertTo-Json|Set-Content (Join-Path $root 'cleanup.json') -Encoding utf8
Get-Content (Join-Path $root 'cleanup.json') -Raw -Encoding utf8
