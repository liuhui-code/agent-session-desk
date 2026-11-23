$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-search-20261123'
$identities=@();foreach($name in @('Copy_A','Copy_B','中文会话')){$identities+=,([IO.File]::ReadAllText((Join-Path $root ('work/'+$name+'.identity.json')))|ConvertFrom-Json)}
foreach($name in @('ui-identities.json','style-ui-identities.json')){$items=[IO.File]::ReadAllText((Join-Path $root $name))|ConvertFrom-Json;foreach($item in $items){$identities+=,$item}}
foreach($folder in @('desktop','style-desktop')){$ready=[IO.File]::ReadAllText((Join-Path $root ($folder+'/ready.json')))|ConvertFrom-Json;$identities+=,@{pid=$ready.pid;filetime=$ready.filetime;path=$ready.path}}
$checked=@();foreach($id in $identities){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.Path -ne $id.path);if(!$ended){throw 'Original registered process remains'};$checked+=,@{pid=$id.pid;filetime=$id.filetime;path=$id.path;original_ended=$ended;pid_reused=[bool]$p}}
$paths=@('search-product.exe','search-product-final.exe','search-ui-tests.exe','search-ui-tests-final.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count){throw 'Exact private image remains'}
$tasks=@('AsdSessionSearchSetup20261123','AsdSessionSearchDesktop20261123','AsdSessionSearchStyleDesktop20261123');foreach($name in $tasks){if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
$clips=@();foreach($folder in @('desktop','style-desktop')){$clip=[IO.File]::ReadAllText((Join-Path $root ($folder+'/clipboard-restored.json')))|ConvertFrom-Json;if(!$clip.restored){throw 'Clipboard text/format restore failed'};$clips+=,@{phase=$folder;text_restored=$clip.text_restored;format_names_restored=$clip.format_names_restored;nontext_content_verified=$false}}
$settings=Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json';$hash=(Get-FileHash $settings).Hash;if($hash -ne '1499307ABEBA68C23F9EB38A4C636F81172475B6B2B535BC0B11DCD9547241F8'){throw 'Settings not restored'}
$audit=@{result='passed';utc=[DateTime]::UtcNow.ToString('o');scope='Known private test images and original registered identities only; no machine-global psmux claim';original_processes=$checked;private_images_alive=$live.Count;tasks_absent=$tasks;clipboard_phases=$clips;settings_sha256=$hash;final_product_sha256=(Get-FileHash (Join-Path $base 'search-product-final.exe')).Hash;final_ui_test_sha256=(Get-FileHash (Join-Path $base 'search-ui-tests-final.exe')).Hash}
$audit|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'final-audit.json') -Encoding utf8
$export=Join-Path $base 'session-search-export';if(Test-Path $export){throw 'Fresh export required'};$null=New-Item -ItemType Directory $export
foreach($file in Get-ChildItem $root -File){if($file.Extension -in @('.json','.log')){Copy-Item $file.FullName (Join-Path $export $file.Name)}}
Copy-Item (Join-Path $root 'work') (Join-Path $export 'work') -Recurse
foreach($folder in @('desktop','style-desktop')){
 $from=Join-Path $root $folder;$to=Join-Path $export $folder;$null=New-Item -ItemType Directory $to
 foreach($name in @('responses','done')){Copy-Item (Join-Path $from $name) (Join-Path $to $name) -Recurse}
 foreach($file in Get-ChildItem $from -File){if($file.Extension -eq '.png' -or $file.Name -in @('ready.json','clipboard-restored.json','driver-error.json')){Copy-Item $file.FullName (Join-Path $to $file.Name)}}
}
Copy-Item (Join-Path $base 'search-setup-outcome.json') (Join-Path $export 'setup-outcome.json')
Add-Type -AssemblyName System.IO.Compression.FileSystem;$zip=Join-Path $base 'session-search-export.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{result=$audit.result;original_processes_ended=$checked.Count;private_images_alive=$live.Count;archive_bytes=(Get-Item $zip).Length;archive_sha256=(Get-FileHash $zip).Hash;settings_sha256=$hash}|ConvertTo-Json
