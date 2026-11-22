$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tabs-ui-20261123'
$dest=Join-Path $root 'collected';New-Item -ItemType Directory -Path $dest|Out-Null
foreach($f in @('go-ui.log','go-ui-run2.log','go-ui-run3.log','go-ui-final.log','binaries.json','binaries-run3.json','binaries-final.json','sessions.json','resize-probe-frames-run3.json','resize-probe-frames-run4.json','terminal-settings.before')){if(Test-Path (Join-Path $root $f)){Copy-Item (Join-Path $root $f) $dest}}
foreach($f in @('paste-probe.json','paste-probe-error.txt','scan-probe.json','scan-probe-error.txt')){if(Test-Path (Join-Path $root "desktop/$f")){Copy-Item (Join-Path $root "desktop/$f") $dest}}
foreach($name in @('desktop','desktop2','native-backend-run3')){if(Test-Path (Join-Path $root $name)){Copy-Item (Join-Path $root $name) (Join-Path $dest $name) -Recurse}}
@{collected=$dest;utc=[DateTime]::UtcNow.ToString('o');app_sha256=(Get-FileHash (Join-Path $base 'tabs-ui-final.exe')).Hash;ui_test_sha256=(Get-FileHash (Join-Path $base 'tabs-ui-tests-final.exe')).Hash;backend_test_sha256=(Get-FileHash (Join-Path $base 'tabs-backend-tests-run3.exe')).Hash}|ConvertTo-Json
