$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-rename-desktop-20261123'
$flow=[IO.File]::ReadAllText((Join-Path $root 'rename-result.json'))|ConvertFrom-Json
$ids=[Collections.Generic.List[object]]::new()
foreach($name in @('Copy_A','Copy_B','中文会话')){$id=[IO.File]::ReadAllText((Join-Path $root ('work/'+$name+'.identity.json')))|ConvertFrom-Json;$ids.Add($id)}
foreach($id in $flow.ui_processes){$ids.Add($id)}
$ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json;$ids.Add(@{pid=$ready.pid;filetime=$ready.filetime;path=$ready.path})
if($ids.Count -ne 7){throw 'Expected exactly 3 shells, 3 UI instances and 1 Terminal'}
$checked=[Collections.Generic.List[object]]::new()
foreach($id in $ids){
 if($id.pid -is [Array] -or $id.path -isnot [string] -or $id.filetime -is [Array]){throw 'Non-scalar identity; no batched audit'}
 $p=Get-Process -Id ([int]$id.pid) -ErrorAction SilentlyContinue
 $ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne [long]$id.filetime -or $p.Path -ne $id.path)
 if(!$ended){throw 'Registered original process remains'}
 $checked.Add(@{pid=[int]$id.pid;filetime=[long]$id.filetime;path=$id.path;original_ended=$ended;pid_reused=[bool]$p})
}
$paths=@('asd-rename.exe','asd-rename-final.exe','rename-session-tests.exe','rename-store-tests.exe','rename-manager-tests.exe','rename-ui-tests.exe','rename-ui-tests-final.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count){throw 'Private image remains'}
$tasks=@('AsdSessionRenameSetup20261123','AsdSessionRenameDesktop20261123');foreach($task in $tasks){if(Get-ScheduledTask $task -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
$settings=Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json';$hash=(Get-FileHash $settings).Hash
if($hash -ne '1499307ABEBA68C23F9EB38A4C636F81172475B6B2B535BC0B11DCD9547241F8'){throw 'Settings not restored'}
$clip=[IO.File]::ReadAllText((Join-Path $root 'desktop/clipboard-restored.json'))|ConvertFrom-Json;if(!$clip.restored){throw 'Clipboard not restored'}
$audit=@{result='passed';utc=[DateTime]::UtcNow.ToString('o');original_processes=$checked.ToArray();private_images_alive=$live.Count;tasks_absent=$tasks;settings_sha256=$hash;clipboard=$clip;corrects='Initial audit enumerated a wrapped UI array as one record; this audit checks all seven scalar PID/FILETIME/path identities separately. No stop, input or setting mutation repeated.'}
$file=Join-Path $root 'final-audit-corrected.json';$audit|ConvertTo-Json -Depth 9|Set-Content $file -Encoding utf8
$export=Join-Path $base 'session-rename-export';Copy-Item $file (Join-Path $export 'final-audit-corrected.json')
Add-Type -AssemblyName System.IO.Compression.FileSystem;$zip=Join-Path $base 'session-rename-export-corrected.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{result='passed';original_processes_ended=$checked.Count;private_images_alive=$live.Count;archive_bytes=(Get-Item $zip).Length;archive_sha256=(Get-FileHash $zip).Hash;settings_sha256=$hash}|ConvertTo-Json
