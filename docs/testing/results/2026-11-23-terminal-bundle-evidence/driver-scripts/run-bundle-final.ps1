$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'terminal-bundle-20261123'
$env:ASD_TERMINAL_BUNDLE_TEST_DIR=Join-Path $root 'bundle'
$env:ASD_PSMUX_TEST_BINARY=Join-Path $base 'history-backend-candidate/psmux.exe'
$env:ASD_PSMUX_TEST_SHA256='ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'
$env:ASD_FIXTURE_TEST_BINARY=Join-Path $base 'history-fixture.exe'
$env:ASD_PRODUCT_TEST_BINARY=Join-Path $base 'bundle-product.exe'
$env:ASD_SHELL_TEST_BINARY=Join-Path $base 'pwsh7/pwsh.exe'
$env:ASD_NATIVE_EVIDENCE_DIR=Join-Path $root 'final-fixture-evidence'
& (Join-Path $base 'bundle-backend-final.test.exe') '-test.v' '-test.run=^TestWindows(TerminalBundle|PsmuxNative)'  '-test.count=1' | Tee-Object (Join-Path $root 'final.log')
$code=$LASTEXITCODE
@{utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;test_sha256=(Get-FileHash (Join-Path $base 'bundle-backend-final.test.exe')).Hash;bundle_files=@(Get-ChildItem $env:ASD_TERMINAL_BUNDLE_TEST_DIR -File|ForEach-Object{@{name=$_.Name;sha256=(Get-FileHash $_.FullName).Hash}})}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'final-result.json') -Encoding utf8
Get-Content (Join-Path $root 'final-result.json') -Raw -Encoding utf8

& (Join-Path $base 'bundle-cli-final.test.exe') '-test.v' '-test.run=^TestTerminalBundle' '-test.count=1' | Tee-Object (Join-Path $root 'cli-final.log')
if($LASTEXITCODE -ne 0){throw 'CLI bundle rejection failed'}
if($code -ne 0){throw 'Native final suite failed'}
