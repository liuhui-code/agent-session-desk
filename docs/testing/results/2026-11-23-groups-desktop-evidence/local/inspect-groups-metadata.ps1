$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'group-workspace-desktop-20261123'
foreach($name in @('sessions.json','flow-failure.json','ui-identities.json')){if(Test-Path (Join-Path $root $name)){$text=[IO.File]::ReadAllText((Join-Path $root $name));$text.Replace($env:USERPROFILE,'<WINDOWS_TEST_HOME>')}}
