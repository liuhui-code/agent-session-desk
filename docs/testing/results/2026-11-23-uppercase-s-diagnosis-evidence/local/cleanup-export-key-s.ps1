$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'sidebar-key-diagnosis-20261123'
$product=Join-Path $base 'asd-sidebar-key-diag.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe'
$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-sidebar-key-diag','--state-dir',(Join-Path $root 'state'))
if((Get-FileHash $product).Hash.ToLowerInvariant() -cne 'f58e80afbef817bfbf7e42528a57607a4e664bf068e2dd417ff1f6b4a7acf277'){throw 'Final product differs'}
$result=[IO.File]::ReadAllText((Join-Path $root 'key-s-result.json'))|ConvertFrom-Json
if($result.result -ne 'passed'){throw 'Flow not passed'}
if($result.caps_before.capsLock -ne $true -or $result.caps_after_control.capsLock -ne $result.caps_before.capsLock){throw 'Original Caps Lock was not restored before closing the window'}
foreach($taskName in @('AsdSidebarKeyDesktop20261123','AsdSidebarKeyContinuation20261123','AsdSidebarKeyContinuation220261123')){$until=[DateTime]::UtcNow.AddSeconds(15);while((Get-ScheduledTask -TaskName $taskName).State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if((Get-ScheduledTask -TaskName $taskName).State -eq 'Running'){throw 'Driver still running; no cleanup or task termination'}}
$terminal=Get-Item (Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe')
if(@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $terminal.FullName}).Count){throw 'Private window still running'}
$owned=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json
$rows=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $rows.Count -ne 1){throw 'Unexpected live set'}
foreach($v in $rows){if(@($owned|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id -and $_.backend_id -eq $v.backend_id}).Count -ne 1){throw 'Unowned target; no stops'}}
$stops=@();foreach($v in $rows){$text=(& $product @g stop $v.session_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Exact owned stop failed'};$stops+=,@{session_id=$v.session_id;run_id=$v.run_id;result=$text}}
$list=(& $product @g list --json|Out-String).Trim();if($LASTEXITCODE -ne 0 -or $list -ne '[]'){throw 'Owned sessions remain'}
$receipt=[IO.File]::ReadAllText((Join-Path $root 'receipt.json'))|ConvertFrom-Json
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json'
if((Get-FileHash $settings).Hash -ne $receipt.settings_test){throw 'Settings changed externally; not overwritten'}
[IO.File]::Copy((Join-Path $root 'settings.before'),$settings,$true)
if((Get-FileHash $settings).Hash -ne $receipt.settings_before){throw 'Settings restore mismatch'}
foreach($taskName in @('AsdSidebarKeyDesktop20261123','AsdSidebarKeyContinuation20261123','AsdSidebarKeyContinuation220261123')){Unregister-ScheduledTask -TaskName $taskName -Confirm:$false}
$clip=[IO.File]::ReadAllText((Join-Path $root 'desktop/clipboard-restored.json'))|ConvertFrom-Json
if(!$clip.restored){throw 'Clipboard text/format restoration failed'}
$identities=@();foreach($name in @('Copy_A')){$identities+=,([IO.File]::ReadAllText((Join-Path $root ('work/'+$name+'.identity.json')))|ConvertFrom-Json)}
foreach($id in $result.ui_processes){$identities+=,$id}
$ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json;$identities+=,@{pid=$ready.pid;filetime=$ready.filetime;path=$ready.path}
if($identities.Count -ne 5){throw 'Expected five scalar navigation identities'}
$audit=@();foreach($id in $identities){if($id.pid -is [Array] -or $id.path -isnot [string]){throw 'Non-scalar style identity'};$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.Path -ne $id.path);if(!$ended){throw 'Registered original process remains'};$audit+=,@{pid=$id.pid;filetime=$id.filetime;path=$id.path;original_ended=$ended;pid_reused=[bool]$p}}
$paths=@('asd-sidebar-key-diag.exe','asd-key-s-instrumented.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count){throw 'Known private image remains'}
$tasks=@('AsdSidebarKeyContinuation220261123','AsdSidebarKeyContinuation20261123','AsdSidebarKeySetup20261123','AsdSidebarKeyDesktop20261123','AsdSidebarKeyLayoutQuery20261123','AsdSidebarKeyCapsBefore20261123','AsdSidebarKeyCapsOff20261123','AsdSidebarKeyCapsCurrent20261123','AsdSidebarKeyCapsFinal20261123','AsdSidebarKeyCapsAftercontrol20261123');foreach($name in $tasks){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');stops=$stops;remaining_sessions=$list;original_processes=$audit;caps_before=$result.caps_before.capsLock;caps_verified_before_window_close=$result.caps_after_control.capsLock;private_images_alive=$live.Count;tasks_absent=$tasks;settings_sha256=(Get-FileHash $settings).Hash;clipboard=$clip;scope='Registered original identities and known private image paths only; no machine-global cleanup claim; owned target and outer launcher PSReadLine history were isolated before input; no user history content archived'}|ConvertTo-Json -Depth 9|Set-Content (Join-Path $root 'final-audit.json') -Encoding utf8
$export=Join-Path $base 'key-s-diagnosis-export';if(Test-Path $export){throw 'Fresh export required'};$null=New-Item -ItemType Directory $export
foreach($f in Get-ChildItem $root -File){if($f.Extension -in @('.json','.jsonl','.log')){Copy-Item $f.FullName (Join-Path $export $f.Name)}}
$to=Join-Path $export 'work';$null=New-Item -ItemType Directory $to
foreach($name in @('Copy_A.identity.json','prompt.ps1','alt-app.ps1')){Copy-Item (Join-Path $root ('work/'+$name)) (Join-Path $to $name)}
$from=Join-Path $root 'desktop';$to=Join-Path $export 'desktop';$null=New-Item -ItemType Directory $to
foreach($name in @('responses','done')){Copy-Item (Join-Path $from $name) (Join-Path $to $name) -Recurse}
foreach($name in @('ready.json','continuation-ready.json','driver-continuation-error.json','clipboard-first-driver-restored.json','clipboard-prior-continuation-restored.json','clipboard-restored.json','baseline-sendkeys-1-public.png','baseline-unicode-1-public.png','baseline-sendkeys-2-public.png','baseline-unicode-2-public.png','baseline-sendkeys-3-public.png','baseline-unicode-3-public.png','sendkeys-1-public.png','unicode-1-public.png','held-shift-1-public.png','caps-on-physical-1-public.png','original-caps-aware-1-public.png','original-terminal-uppercase-S-public.png','original-settings-with-draft-public.png')){Copy-Item (Join-Path $from $name) (Join-Path $to $name)}
foreach($name in @('ui-v1-before.toml')){Copy-Item (Join-Path $root $name) (Join-Path $export $name)}
Copy-Item (Join-Path $root 'state/ui.toml') (Join-Path $export 'ui-final.toml')
Copy-Item (Join-Path $base 'key-s-setup-outcome.json') (Join-Path $export 'setup-outcome.json')
Add-Type -AssemblyName System.IO.Compression.FileSystem;$zip=Join-Path $base 'key-s-diagnosis-export.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{result='passed';original_processes_ended=$audit.Count;private_images_alive=$live.Count;archive_bytes=(Get-Item $zip).Length;archive_sha256=(Get-FileHash $zip).Hash;settings_sha256=(Get-FileHash $settings).Hash}|ConvertTo-Json
