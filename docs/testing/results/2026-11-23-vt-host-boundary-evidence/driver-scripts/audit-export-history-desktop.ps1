$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'history-desktop-20261123'
$ids=@();foreach($n in @('Copy_A','Copy_B')){$ids+=,([IO.File]::ReadAllText((Join-Path $root ('work/'+$n+'.identity.json')))|ConvertFrom-Json)}
$ids+=,([IO.File]::ReadAllText((Join-Path $root 'work/alt-started.json'))|ConvertFrom-Json)
$uis=[IO.File]::ReadAllText((Join-Path $root 'ui-identities.json'))|ConvertFrom-Json;foreach($id in $uis){$ids+=,$id}
$ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json;$ids+=,@{pid=$ready.pid;filetime=$ready.filetime;path=$ready.path}
$checked=@();foreach($id in $ids){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.Path -ne $id.path);if(!$ended){throw 'Original registered process remains'};$checked+=,@{pid=$id.pid;filetime=$id.filetime;path=$id.path;original_ended=$ended;pid_reused=[bool]$p}}
if(Get-ScheduledTask AsdHistoryDesktop20261123 -ErrorAction SilentlyContinue){throw 'Desktop task remains'}
$paths=@((Join-Path $base 'copy-final2-product.exe'),(Join-Path $base 'history-backend-candidate/psmux.exe'),(Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe'))
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count){throw 'Exact owned image remains'}
$clip=[IO.File]::ReadAllText((Join-Path $root 'desktop/clipboard-restored.json'))|ConvertFrom-Json;if(!$clip.restored){throw 'Clipboard text/format restore failed'}
$settings=Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json';$hash=(Get-FileHash $settings).Hash
if($hash -ne '1499307ABEBA68C23F9EB38A4C636F81172475B6B2B535BC0B11DCD9547241F8'){throw 'Settings restore failed'}
$obs=[IO.File]::ReadAllText((Join-Path $root 'native-observations.json'))|ConvertFrom-Json
$audit=@{result='passed';utc=[DateTime]::UtcNow.ToString('o');original_processes=$checked;exact_owned_images_alive=$live.Count;task_absent=$true;clipboard_text_restored=$clip.text_restored;clipboard_format_names_restored=$clip.format_names_restored;clipboard_nontext_content_verified=$false;settings_sha256=$hash;commands=@(Get-ChildItem (Join-Path $root 'desktop/responses') -Filter '*.json').Count;product_reused=$true;go_tests_rerun=$false}
$audit|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'final-audit.json') -Encoding utf8
$export=Join-Path $base 'history-desktop-export';if(Test-Path $export){throw 'Fresh export required'};New-Item -ItemType Directory $export|Out-Null
foreach($name in @('receipt.json','sessions.json','ui-identities.json','failure-inspection.json','native-hint-inspection.json','progress-inspection.json','flow-failure.json','lifecycle-failure.json','lifecycle-result.json','lifecycle-observations.json','native-observations.json','cleanup.json','final-audit.json','resume-guard.json')){$p=Join-Path $root $name;if(Test-Path $p){Copy-Item $p (Join-Path $export $name)}}
Copy-Item (Join-Path $root 'work') (Join-Path $export 'work') -Recurse
$desktop=Join-Path $export 'desktop';New-Item -ItemType Directory $desktop|Out-Null
foreach($n in @('responses','done')){Copy-Item (Join-Path (Join-Path $root 'desktop') $n) (Join-Path $desktop $n) -Recurse}
foreach($p in Get-ChildItem (Join-Path $root 'desktop') -File){if($p.Extension -eq '.png' -or $p.Name -in @('ready.json','clipboard-restored.json','driver-error.json')){Copy-Item $p.FullName (Join-Path $desktop $p.Name)}}
Add-Type -AssemblyName System.IO.Compression.FileSystem;$zip=Join-Path $base 'history-desktop-export.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{result=$audit.result;original_processes_ended=$checked.Count;private_images_alive=$live.Count;commands=$audit.commands;settings_sha256=$hash;clipboard_text_restored=$clip.text_restored;clipboard_format_names_restored=$clip.format_names_restored;archive_sha256=(Get-FileHash $zip).Hash;archive_bytes=(Get-Item $zip).Length}|ConvertTo-Json
