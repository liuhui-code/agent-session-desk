$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'history-tool-candidate1';$exe=Join-Path $base 'psmux-history-candidate.exe'
$files=@(Get-ChildItem $root -Recurse -File|ForEach-Object{@{name=[string]$_.Name;size=$_.Length;relative=[string]$_.FullName.Substring($root.Length)}})
$logs=@(Get-ChildItem $root -Recurse -File -Filter '*.log'|Where-Object{$_.Length -le 64KB}|ForEach-Object{@{name=[string]$_.Name;text=[IO.File]::ReadAllText($_.FullName)}})
$identities=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $exe -or ($_.CommandLine -and $_.CommandLine.Contains((Join-Path $root 'history-fixture.ps1')))}|ForEach-Object{$p=Get-Process -Id $_.ProcessId;@{pid=$p.Id;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}})
$r=@{version=(& $exe -V|Out-String);files=$files;logs=$logs;matching_test_processes=$identities}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'startup-inspection.json') -Encoding utf8;$r|ConvertTo-Json -Depth 8
