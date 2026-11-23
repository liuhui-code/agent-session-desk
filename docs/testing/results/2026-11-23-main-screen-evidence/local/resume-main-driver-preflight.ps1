$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-parity-desktop-20261123'
$t=Get-ScheduledTask 'AsdMainParityDesktop20261123';$i=Get-ScheduledTaskInfo $t
if($t.State -eq 'Running' -or $i.LastTaskResult -ne 1){throw 'Original driver terminal outcome not as observed'}
$images=@((Join-Path $base 'asd-main-parity.exe'),(Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe'))
if(@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $images}).Count){throw 'Unexpected original UI/window exists'}
foreach($d in @('responses','done')){if(@(Get-ChildItem (Join-Path $root ('desktop/'+$d)) -File).Count){throw 'Original driver processed an action'}}
$c=[IO.File]::ReadAllText((Join-Path $root 'desktop/commands/001.json'))|ConvertFrom-Json
if($c.kind -ne 'state'){throw 'Unexpected queued action'}
$ids=@();foreach($n in @('Copy_A','Copy_B')){$v=[IO.File]::ReadAllText((Join-Path $root ('work/'+$n+'.identity.json')))|ConvertFrom-Json;$p=Get-Process -Id $v.pid;if($p.Path -ne $v.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $v.filetime){throw 'Original shell changed'};$ids+=,$v}
foreach($name in @('main-parity-desktop-driver.ps1','main-parity-flow.ps1')){$tokens=$null;$errors=$null;$null=[Management.Automation.Language.Parser]::ParseInput([IO.File]::ReadAllText((Join-Path $base $name),[Text.Encoding]::UTF8),[ref]$tokens,[ref]$errors);if($errors.Count){throw 'Corrected script syntax error'}}
Copy-Item (Join-Path $root 'flow-failure.json') (Join-Path $root 'first-flow-failure.json')
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');old_task_state=[string]$t.State;old_task_result=$i.LastTaskResult;accepted_actions=0;queued_readonly=1;replayed_mutations=0;original_shells=$ids;new_command_directory='desktop-run2';clipboard_scope='First driver had no captured backup/restoration result; corrected driver fails closed before any window/input when backup unavailable'}|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'driver-resume-preflight.json') -Encoding utf8
Get-Content (Join-Path $root 'driver-resume-preflight.json') -Raw -Encoding utf8
