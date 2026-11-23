$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-rename-surface-20261123'
$product=Join-Path $base 'asd-rename-surface.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe'
$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-rename-surface','--state-dir',(Join-Path $root 'state'))
if((Get-FileHash $product).Hash.ToLowerInvariant() -cne '8c8aea2a17f94f593dd99a12f55d3f3c3e15db36b1201090243bab393fa573fd'){throw 'Final product differs'}
$result=[IO.File]::ReadAllText((Join-Path $root 'style-result.json'))|ConvertFrom-Json
if($result.result -ne 'passed'){throw 'Flow not passed'}
$taskName='AsdSessionRenameSurfaceDesktop20261123';$until=[DateTime]::UtcNow.AddSeconds(15)
while((Get-ScheduledTask -TaskName $taskName).State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if((Get-ScheduledTask -TaskName $taskName).State -eq 'Running'){throw 'Driver still running; no cleanup or task termination'}
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
Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
$clip=[IO.File]::ReadAllText((Join-Path $root 'desktop/clipboard-restored.json'))|ConvertFrom-Json
if(!$clip.restored){throw 'Clipboard text/format restoration failed'}
$identities=@();foreach($name in @('Copy_A')){$identities+=,([IO.File]::ReadAllText((Join-Path $root ('work/'+$name+'.identity.json')))|ConvertFrom-Json)}
foreach($id in $result.ui_processes){$identities+=,$id}
$ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json;$identities+=,@{pid=$ready.pid;filetime=$ready.filetime;path=$ready.path}
if($identities.Count -ne 3){throw 'Expected three scalar style identities'}
$audit=@();foreach($id in $identities){if($id.pid -is [Array] -or $id.path -isnot [string]){throw 'Non-scalar style identity'};$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.Path -ne $id.path);if(!$ended){throw 'Registered original process remains'};$audit+=,@{pid=$id.pid;filetime=$id.filetime;path=$id.path;original_ended=$ended;pid_reused=[bool]$p}}
$paths=@('asd-rename.exe','asd-rename-surface.exe','rename-session-tests.exe','rename-store-tests.exe','rename-manager-tests.exe','rename-ui-tests.exe','rename-ui-tests-final.exe','rename-ui-tests-surface.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count){throw 'Known private image remains'}
$tasks=@('AsdSessionRenameSurfaceSetup20261123','AsdSessionRenameSurfaceDesktop20261123');foreach($name in $tasks){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');stops=$stops;remaining_sessions=$list;original_processes=$audit;private_images_alive=$live.Count;tasks_absent=$tasks;settings_sha256=(Get-FileHash $settings).Hash;clipboard=$clip;scope='Registered original identities and known private image paths only; no machine-global cleanup claim; private PowerShell history restoration not verified'}|ConvertTo-Json -Depth 9|Set-Content (Join-Path $root 'final-audit.json') -Encoding utf8
$export=Join-Path $base 'session-rename-surface-export';if(Test-Path $export){throw 'Fresh export required'};$null=New-Item -ItemType Directory $export
foreach($f in Get-ChildItem $root -File){if($f.Extension -in @('.json','.log')){Copy-Item $f.FullName (Join-Path $export $f.Name)}}
$to=Join-Path $export 'work';$null=New-Item -ItemType Directory $to
foreach($name in @('Copy_A.identity.json','prompt.ps1','alt-app.ps1')){Copy-Item (Join-Path $root ('work/'+$name)) (Join-Path $to $name)}
$from=Join-Path $root 'desktop';$to=Join-Path $export 'desktop';$null=New-Item -ItemType Directory $to
foreach($name in @('responses','done')){Copy-Item (Join-Path $from $name) (Join-Path $to $name) -Recurse}
foreach($name in @('ready.json','clipboard-restored.json','rename-surface-120-public.png','rename-surface-63-public.png','rename-surface-49-public.png','rename-surface-error-49-public.png')){Copy-Item (Join-Path $from $name) (Join-Path $to $name)}
Copy-Item (Join-Path $base 'rename-surface-setup-outcome.json') (Join-Path $export 'setup-outcome.json')
Add-Type -AssemblyName System.IO.Compression.FileSystem;$zip=Join-Path $base 'session-rename-surface-export.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{result='passed';original_processes_ended=$audit.Count;private_images_alive=$live.Count;archive_bytes=(Get-Item $zip).Length;archive_sha256=(Get-FileHash $zip).Hash;settings_sha256=(Get-FileHash $settings).Hash}|ConvertTo-Json
