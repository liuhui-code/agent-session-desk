$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-rename-desktop-20261123'
$product=Join-Path $base 'asd-rename-final.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe'
$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-session-rename','--state-dir',(Join-Path $root 'state'))
if((Get-FileHash $product).Hash.ToLowerInvariant() -cne 'e39e042fec454d5e306e7819a8783a38ea5d5f2c229fe463b8962d974bfc4c24'){throw 'Final product differs'}
$result=[IO.File]::ReadAllText((Join-Path $root 'rename-result.json'))|ConvertFrom-Json
if($result.result -ne 'passed'){throw 'Flow not passed'}
$taskName='AsdSessionRenameDesktop20261123';$until=[DateTime]::UtcNow.AddSeconds(15)
while((Get-ScheduledTask -TaskName $taskName).State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if((Get-ScheduledTask -TaskName $taskName).State -eq 'Running'){throw 'Driver still running; no cleanup or task termination'}
$terminal=Get-Item (Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe')
if(@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $terminal.FullName}).Count){throw 'Private window still running'}
$owned=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json
$b=@($owned|Where-Object{$_.title -ceq 'Copy B'})[0]
$record=(& $product @g record $b.session_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Renamed real Run cannot record output'}
$recorded=$record|ConvertFrom-Json
$output=(& $product @g output $b.run_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Renamed real Run cannot read output'}
$records=$output|ConvertFrom-Json
if($records.Count -ne 1 -or $records[0].run_id -ne $b.run_id -or $records[0].session_id -ne $b.session_id -or $records[0].complete -or $records[0].text -notmatch 'RENAME_B:17392'){throw 'Recorded output lost original target or source boundary'}
@{result='passed';record=$recorded;output=$records}|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'renamed-output-result.json') -Encoding utf8
$rows=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $rows.Count -ne 3){throw 'Unexpected live set'}
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
$identities=@();foreach($name in @('Copy_A','Copy_B','中文会话')){$identities+=,([IO.File]::ReadAllText((Join-Path $root ('work/'+$name+'.identity.json')))|ConvertFrom-Json)}
foreach($id in @([IO.File]::ReadAllText((Join-Path $root 'ui-identities.json'))|ConvertFrom-Json)){$identities+=,$id}
$ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json;$identities+=,@{pid=$ready.pid;filetime=$ready.filetime;path=$ready.path}
$audit=@();foreach($id in $identities){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.Path -ne $id.path);if(!$ended){throw 'Registered original process remains'};$audit+=,@{pid=$id.pid;filetime=$id.filetime;path=$id.path;original_ended=$ended;pid_reused=[bool]$p}}
$paths=@('asd-rename.exe','asd-rename-final.exe','rename-session-tests.exe','rename-store-tests.exe','rename-manager-tests.exe','rename-ui-tests.exe','rename-ui-tests-final.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count){throw 'Known private image remains'}
$tasks=@('AsdSessionRenameSetup20261123','AsdSessionRenameDesktop20261123');foreach($name in $tasks){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');stops=$stops;remaining_sessions=$list;original_processes=$audit;private_images_alive=$live.Count;tasks_absent=$tasks;settings_sha256=(Get-FileHash $settings).Hash;clipboard=$clip;scope='Registered original identities and known private image paths only; no machine-global cleanup claim; private PowerShell history restoration not verified'}|ConvertTo-Json -Depth 9|Set-Content (Join-Path $root 'final-audit.json') -Encoding utf8
$export=Join-Path $base 'session-rename-export';if(Test-Path $export){throw 'Fresh export required'};$null=New-Item -ItemType Directory $export
foreach($f in Get-ChildItem $root -File){if($f.Extension -in @('.json','.log')){Copy-Item $f.FullName (Join-Path $export $f.Name)}}
$to=Join-Path $export 'work';$null=New-Item -ItemType Directory $to
foreach($name in @('Copy_A.identity.json','Copy_B.identity.json','中文会话.identity.json','prompt.ps1','alt-app.ps1')){Copy-Item (Join-Path $root ('work/'+$name)) (Join-Path $to $name)}
$from=Join-Path $root 'desktop';$to=Join-Path $export 'desktop';$null=New-Item -ItemType Directory $to
foreach($name in @('responses','done')){Copy-Item (Join-Path $from $name) (Join-Path $to $name) -Recurse}
foreach($name in @('ready.json','clipboard-restored.json','rename-background-120-public.png','rename-prefill-63-public.png','rename-prefill-49-public.png','rename-duplicate-identities-120-public.png')){Copy-Item (Join-Path $from $name) (Join-Path $to $name)}
Copy-Item (Join-Path $base 'rename-setup-outcome.json') (Join-Path $export 'setup-outcome.json')
Add-Type -AssemblyName System.IO.Compression.FileSystem;$zip=Join-Path $base 'session-rename-export.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{result='passed';original_processes_ended=$audit.Count;private_images_alive=$live.Count;archive_bytes=(Get-Item $zip).Length;archive_sha256=(Get-FileHash $zip).Hash;settings_sha256=(Get-FileHash $settings).Hash}|ConvertTo-Json
