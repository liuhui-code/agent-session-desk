$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'visual-replay-utf8-20261123'
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName}).Count -ne 0){throw 'Private window still running'}
foreach($name in @('AsdVisualReplayUtf820261123')){$task=Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue;if($task -and $task.State -eq 'Running'){throw 'Owned task still running'}}
$r=[IO.File]::ReadAllText((Join-Path $root 'receipt.json'))|ConvertFrom-Json
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';if((Get-FileHash $settings).Hash -ne $r.settings_test){throw 'Settings changed externally'}
[IO.File]::Copy((Join-Path $root 'settings.before'),$settings,$true);if((Get-FileHash $settings).Hash -ne $r.settings_before){throw 'Restore mismatch'}
foreach($name in @('AsdVisualReplayUtf820261123')){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){Unregister-ScheduledTask -TaskName $name -Confirm:$false}}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');reference_only=$true;private_windows=0;settings_restored=(Get-FileHash $settings).Hash}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'cleanup.json') -Encoding utf8
Get-Content (Join-Path $root 'cleanup.json') -Raw -Encoding utf8
