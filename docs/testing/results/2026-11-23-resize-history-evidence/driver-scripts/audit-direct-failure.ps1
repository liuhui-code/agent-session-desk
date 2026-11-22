$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/history-direct-control'
$i=[IO.File]::ReadAllText((Join-Path $root 'identity.json'))|ConvertFrom-Json;$p=Get-Process -Id $i.pid -ErrorAction SilentlyContinue
$r=@{identity=$i;original_ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $i.filetime -or $p.Path -ne $i.path);registry_entry_count=@(Get-ChildItem (Join-Path $root 'runtime/asd-history-direct-control')).Count;matching_fixture_processes=@(Get-CimInstance Win32_Process|Where-Object{$_.CommandLine -and $_.CommandLine.Contains((Join-Path $root 'history-fixture.ps1'))}).Count}
$r|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'failure-audit.json') -Encoding utf8;$r|ConvertTo-Json -Depth 6
if(!$r.original_ended -or $r.matching_fixture_processes -ne 0){throw 'Owned fixture remains'}
