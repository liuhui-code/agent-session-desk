$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$paths=@()
foreach($root in @('instance-product-20261126','instance-final-product-20261126','instance-final-product-r2-20261126','instance-final-product-r3-20261126')){foreach($n in @('asd-instance-clean.exe','asd-instance-r2.exe','asd-instance-r3.exe')){$paths+=Join-Path (Join-Path $base $root) $n}}
foreach($n in @('instance.test.exe','instance-final.test.exe','instance-recovery.test.exe','instance-r2.test.exe','instance-r3.test.exe','pet-instance-clean.test.exe','pet-instance-r2.test.exe','pet-instance-r3.test.exe','cli-instance-r3.test.exe','asd-fixture-integrated.exe')){$paths+=Join-Path $base $n}
$counts=@{};foreach($p in Get-Process -ErrorAction SilentlyContinue){try{if($p.Path-in $paths){$n=[IO.Path]::GetFileName($p.Path);if(!$counts.ContainsKey($n)){$counts[$n]=0};$counts[$n]++}}catch{}}
$tasks=@('AsdInstanceStep120261126','AsdInstanceComponents20261126','AsdInstanceProductProbe20261126','AsdInstanceComponentsR320261126','AsdInstanceDesktop20261126','AsdInstanceDesktopR220261126','AsdInstanceDesktopR320261126')
$present=@($tasks|Where-Object{Get-ScheduledTask -TaskName $_ -ErrorAction SilentlyContinue})
$out=Join-Path $base 'instance-desktop-r3-20261126';$observers=@(Get-ChildItem $out -Filter 'inspect-*.json'|Sort-Object Name|ForEach-Object{Get-Content -Raw $_.FullName|ConvertFrom-Json})
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
@{utc=[DateTime]::UtcNow.ToString('o');owned_image_counts=$counts;tasks_present=$present;observer_count=$observers.Count;observers=$observers;host_fingerprints=@(@{kind='terminal';sha256=(Get-FileHash $hostInfo.terminal.path).Hash.ToLowerInvariant();version=(Get-Item $hostInfo.terminal.path).VersionInfo.FileVersion},@{kind='shell';sha256=(Get-FileHash $hostInfo.shell.path).Hash.ToLowerInvariant();version=(Get-Item $hostInfo.shell.path).VersionInfo.FileVersion});os_build=[Environment]::OSVersion.Version.ToString()}|ConvertTo-Json -Depth 6
