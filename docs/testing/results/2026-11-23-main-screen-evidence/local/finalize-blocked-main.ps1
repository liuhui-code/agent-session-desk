$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-parity-desktop-20261123';$product=Join-Path $base 'asd-main-parity.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe'
$close=[IO.File]::ReadAllText((Join-Path $root 'blocked-ui-close.json'))|ConvertFrom-Json
if($close.remaining_registered -ne 0){throw 'Originals remain'}
$tasks=@('AsdMainParityDesktop20261123','AsdMainParityClose20261123')
foreach($n in $tasks){$t=Get-ScheduledTask $n;$until=[DateTime]::UtcNow.AddSeconds(5);while($t.State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100;$t=Get-ScheduledTask $n};if($t.State -eq 'Running'){throw 'Owned task still running'}}
$ids=([IO.File]::ReadAllText((Join-Path $root 'blocked-cleanup-identities.json'))|ConvertFrom-Json).identities;$audit=@()
foreach($id in $ids){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$ended=(!$p -or $p.Path -ne $id.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime);if(!$ended){throw 'Registered process remains'};$audit+=,@{kind=$id.kind;pid=$id.pid;filetime=$id.filetime;path=$id.path;original_ended=$ended;pid_reused=[bool]$p}}
$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-main-parity','--state-dir',(Join-Path $root 'state'))
$list=(& $product @g list --json|Out-String).Trim();if($LASTEXITCODE -ne 0 -or $list -cne '[]'){throw 'Owned Sessions remain'}
$paths=@('asd-main-parity.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count){throw 'Known private image remains'}
$receipt=[IO.File]::ReadAllText((Join-Path $root 'receipt.json'))|ConvertFrom-Json;$settings=Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json'
if((Get-FileHash $settings).Hash -cne $receipt.settings_test){throw 'Test settings changed externally; not overwritten'}
[IO.File]::Copy((Join-Path $root 'settings.before'),$settings,$true);if((Get-FileHash $settings).Hash -cne $receipt.settings_before){throw 'Settings restore differs'}
foreach($n in $tasks){Unregister-ScheduledTask $n -Confirm:$false}
$absent=@('AsdMainParitySetup20261123','AsdMainParityDesktop20261123','AsdMainParityRead20261123','AsdMainParityClose20261123');foreach($n in $absent){if(Get-ScheduledTask $n -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
$pref=[IO.File]::ReadAllBytes((Join-Path $root 'state/ui.toml'));$seed=[IO.File]::ReadAllBytes((Join-Path $root 'ui-v3-before.toml'));if([Convert]::ToBase64String($pref) -cne [Convert]::ToBase64String($seed)){throw 'UI prefs changed despite no actions'}
@{result='cleanup_passed_desktop_blocked';utc=[DateTime]::UtcNow.ToString('o');accepted_ui_actions=0;original_processes=$audit;private_images_alive=$live.Count;remaining_sessions=$list;tasks_absent=$absent;close_window_requested=$close.close_window_requested;forced_count=@($close.forced).Count;settings_sha256=(Get-FileHash $settings).Hash;preferences_unchanged=$true;first_clipboard_restoration='unknown: backup failed and first cleanup had no restoration result';run2_clipboard_operations=0;desktop_result='blocked: foreground guard refused input before ready';scope='Cleanup of exact five registered originals, not normal manager q or desktop acceptance'}|ConvertTo-Json -Depth 9|Set-Content (Join-Path $root 'final-audit.json') -Encoding utf8
$export=Join-Path $base 'main-parity-blocked-export';if(Test-Path $export){throw 'Fresh export required'};$null=New-Item -ItemType Directory $export
foreach($p in Get-ChildItem $root -File){if($p.Extension -in @('.json','.log')){Copy-Item $p.FullName (Join-Path $export $p.Name)}}
foreach($n in @('desktop','desktop-run2')){Copy-Item (Join-Path $root $n) (Join-Path $export $n) -Recurse}
$work=Join-Path $export 'work';$null=New-Item -ItemType Directory $work;foreach($n in @('Copy_A.identity.json','Copy_B.identity.json','prompt.ps1')){Copy-Item (Join-Path $root ('work/'+$n)) (Join-Path $work $n)}
Copy-Item (Join-Path $root 'ui-v3-before.toml') (Join-Path $export 'ui-v3-before.toml');Copy-Item (Join-Path $root 'state/ui.toml') (Join-Path $export 'ui-final.toml')
Copy-Item (Join-Path $base 'main-parity-tests-final-20261123') (Join-Path $export 'native-tests') -Recurse
Copy-Item (Join-Path $base 'main-parity-setup-outcome.json') (Join-Path $export 'setup-outcome.json');Copy-Item (Join-Path $base 'main-parity-desktop-probe.json') (Join-Path $export 'desktop-probe.json')
Add-Type -AssemblyName System.IO.Compression.FileSystem;$zip=Join-Path $base 'main-parity-blocked-export.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{result='cleanup_passed_desktop_blocked';originals_ended=$audit.Count;forced_count=@($close.forced).Count;private_images_alive=$live.Count;archive_bytes=(Get-Item $zip).Length;archive_sha256=(Get-FileHash $zip).Hash;settings_sha256=(Get-FileHash $settings).Hash}|ConvertTo-Json
