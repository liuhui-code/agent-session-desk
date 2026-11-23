$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'terminal-bundle-20261123'
$paths=@('bundle-product.exe','bundle-backend-red.test.exe','bundle-backend-green.test.exe','bundle-backend-flow.test.exe','bundle-backend-final.test.exe','bundle-cli-final.test.exe','history-fixture.exe','history-backend-candidate/psmux.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths}|ForEach-Object{@{pid=$_.ProcessId;path=$_.ExecutablePath}})
if($live.Count){throw 'Exact private test image remains'}
$ids=@();foreach($folder in @('flow-fixture-evidence','final-fixture-evidence')){
 foreach($file in @(Get-ChildItem (Join-Path $root $folder) -Recurse -Filter root.json)){
  $id=[IO.File]::ReadAllText($file.FullName)|ConvertFrom-Json;$p=Get-Process -Id $id.PID -ErrorAction SilentlyContinue;$ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.Created -or $p.Path -ne $id.Path)
  if(!$ended){throw 'Original bundle flow root remains'};$ids+=,@{pid=$id.PID;filetime=$id.Created;path=$id.Path;original_ended=$ended}
 }
}
$tasks=@('AsdBundleDesktop20261123','AsdBundleUserDesktop20261123','AsdBundlePermissionProbe20261123','AsdBundleUserSetup20261123')
foreach($name in $tasks){if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
$settings=Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json';$hash=(Get-FileHash $settings).Hash
if($hash -ne '1499307ABEBA68C23F9EB38A4C636F81172475B6B2B535BC0B11DCD9547241F8'){throw 'Settings not restored'}
$audit=@{utc=[DateTime]::UtcNow.ToString('o');result='passed';scope='Only known private test images and recorded original identities, not machine-global psmux';exact_private_images_alive=$live.Count;original_roots=$ids;tasks_absent=$tasks;settings_sha256=$hash;product_sha256=(Get-FileHash (Join-Path $base 'bundle-product.exe')).Hash;tests_sha256=(Get-FileHash (Join-Path $base 'bundle-backend-final.test.exe')).Hash;cli_tests_sha256=(Get-FileHash (Join-Path $base 'bundle-cli-final.test.exe')).Hash}
$audit|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'final-audit.json') -Encoding utf8
$export=Join-Path $base 'terminal-bundle-export';if(Test-Path $export){throw 'Fresh export required'};$null=New-Item -ItemType Directory $export
foreach($file in Get-ChildItem $root -File){if($file.Extension -in @('.json','.log')){Copy-Item $file.FullName (Join-Path $export $file.Name)}}
foreach($folder in @('red-fixture-evidence','green-fixture-evidence','flow-fixture-evidence','final-fixture-evidence')){if(Test-Path (Join-Path $root $folder)){Copy-Item (Join-Path $root $folder) (Join-Path $export $folder) -Recurse}}
Copy-Item (Join-Path $base 'bundle-user-setup-outcome.json') (Join-Path $export 'desktop-user-setup-outcome.json')
Add-Type -AssemblyName System.IO.Compression.FileSystem;$zip=Join-Path $base 'terminal-bundle-export.zip';[IO.Compression.ZipFile]::CreateFromDirectory($export,$zip)
@{result=$audit.result;original_roots=$ids.Count;private_images_alive=$live.Count;archive_bytes=(Get-Item $zip).Length;archive_sha256=(Get-FileHash $zip).Hash}|ConvertTo-Json
