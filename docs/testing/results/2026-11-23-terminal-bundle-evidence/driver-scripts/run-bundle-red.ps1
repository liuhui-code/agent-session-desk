$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'terminal-bundle-20261123'
if(Test-Path $root){throw 'Fresh root required'}
$null=New-Item -ItemType Directory $root
$bundle=Join-Path $root 'bundle';$null=New-Item -ItemType Directory $bundle
Copy-Item (Join-Path $base 'history-backend-candidate/psmux.exe') (Join-Path $bundle 'psmux.exe')
foreach($name in @('conpty.dll','OpenConsole.exe')){Copy-Item (Join-Path $base ('f6-reader-20261123/candidate-native/'+$name)) (Join-Path $bundle $name)}
$env:ASD_TERMINAL_BUNDLE_TEST_DIR=$bundle
$env:ASD_PSMUX_TEST_BINARY=Join-Path $base 'history-backend-candidate/psmux.exe'
$env:ASD_PSMUX_TEST_SHA256='ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'
$env:ASD_FIXTURE_TEST_BINARY=Join-Path $base 'history-fixture.exe'
$env:ASD_NATIVE_EVIDENCE_DIR=Join-Path $root 'red-fixture-evidence'
& (Join-Path $base 'bundle-backend-red.test.exe') -test.v -test.run '^TestWindowsTerminalBundleRejectsSystemHostedExistingRun$' -test.count=1 2>&1|Tee-Object (Join-Path $root 'red.log')
$code=$LASTEXITCODE
@{utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;test_sha256=(Get-FileHash (Join-Path $base 'bundle-backend-red.test.exe')).Hash;bundle_files=@(Get-ChildItem $bundle -File|ForEach-Object{@{name=$_.Name;sha256=(Get-FileHash $_.FullName).Hash}})}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'red-result.json') -Encoding utf8
Get-Content (Join-Path $root 'red-result.json') -Raw -Encoding utf8
