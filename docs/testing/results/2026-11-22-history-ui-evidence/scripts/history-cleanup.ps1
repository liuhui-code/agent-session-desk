$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'm1-history-ui-run2'
$terminal=(Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1).FullName
$p=Get-Process -Id 18992 -ErrorAction SilentlyContinue
if($p -and $p.Path -eq $terminal -and $p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq 134398100370036003){throw 'Original owned terminal still running after finish'}
$product=Join-Path $base 'asd-history.exe'
if(@(Get-Process 'asd-history' -ErrorAction SilentlyContinue|Where-Object {$_.Path -eq $product}).Count -ne 0){throw 'Candidate UI still running'}
$name='AsdDesktopHistory20261122'
$task=Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue
if($task -and $task.State -eq 'Running'){throw 'Driver task still running'}
if($task){Unregister-ScheduledTask -TaskName $name -Confirm:$false}
$settings=Join-Path (Split-Path $terminal) 'settings/settings.json';$backup=Join-Path $root 'terminal-settings.before'
$config=[IO.File]::ReadAllText((Join-Path $root 'terminal-config.json'))|ConvertFrom-Json
$currentHash=(Get-FileHash $settings).Hash
if($currentHash -notin @($config.settings_test_sha256,$config.settings_before_sha256)){throw 'Private settings changed externally; restore refused'}
if((Get-FileHash $backup).Hash -ne $config.settings_before_sha256){throw 'Settings backup mismatch'}
if($currentHash -eq $config.settings_test_sha256){[IO.File]::Copy($backup,$settings,$true)}
if((Get-FileHash $settings).Hash -ne $config.settings_before_sha256){throw 'Restore failed'}
$psmux=Join-Path $base 'backend/psmux.exe'
$g=@('--backend',$psmux,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-history-ui')
$list=(& $product @g list --json|Out-String)
if($LASTEXITCODE -ne 0 -or $list.Trim() -ne '[]'){throw 'Owned backend list not empty'}
$history=(& $product @g history|Out-String)
if($LASTEXITCODE -ne 0){throw 'Post UI history failed'}
$before=[IO.File]::ReadAllText((Join-Path $root 'history.json'))|ConvertFrom-Json
$after=$history|ConvertFrom-Json
if(($before|ConvertTo-Json -Depth 10 -Compress) -ne ($after|ConvertTo-Json -Depth 10 -Compress)){throw 'Viewing history changed persisted evidence'}
[IO.File]::WriteAllText((Join-Path $root 'history-after-ui.json'),$history)
@{utc=[DateTime]::UtcNow.ToString('o');terminal_owned_identity_ended=$true;candidate_ui_ended=$true;task_removed=$true;settings_restored=$true;settings_sha256=(Get-FileHash $settings).Hash;backend_list=($list|ConvertFrom-Json);history_unchanged=$true;terminal_sha256=(Get-FileHash $terminal).Hash;psmux_version=(& $psmux -V|Out-String).Trim();powershell_version=$PSVersionTable.PSVersion.ToString()}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'cleanup.json') -Encoding utf8
'Owned window/task cleanup, settings restoration and unchanged history passed.'
