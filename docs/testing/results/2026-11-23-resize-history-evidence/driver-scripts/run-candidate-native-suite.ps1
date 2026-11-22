$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'history-candidate-native-suite'
if(Test-Path $root){throw 'Fresh root required'};New-Item -ItemType Directory $root|Out-Null
$env:ASD_PSMUX_TEST_BINARY=Join-Path $base 'history-backend-candidate/psmux.exe';$env:ASD_PSMUX_TEST_SHA256='ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'
$env:ASD_FIXTURE_TEST_BINARY=Join-Path $base 'history-fixture.exe';$env:ASD_PRODUCT_TEST_BINARY=Join-Path $base 'copy-final2-product.exe';$env:ASD_SHELL_TEST_BINARY=Join-Path $base 'pwsh7/pwsh.exe';$env:ASD_NATIVE_EVIDENCE_DIR=Join-Path $root 'fixture-evidence'
$test=Join-Path $base 'history-backend-tests.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -ne '9e0ff3d77fa12c50e18e87886015ab95eaf3695d7381dd895f4d16e0a6301663'){throw 'Test binary differs'}
if((Get-FileHash $env:ASD_FIXTURE_TEST_BINARY).Hash.ToLowerInvariant() -ne '470ff13355fe4360623a70cf75d28435b8c34d3b1d638c02591523b08efc2442'){throw 'Fixture binary differs'}
& $test '-test.v' '-test.run=^TestWindowsPsmuxNative' *> (Join-Path $root 'native-tests.log');$code=$LASTEXITCODE
$r=@{utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;psmux_sha256=(Get-FileHash $env:ASD_PSMUX_TEST_BINARY).Hash;test_sha256=(Get-FileHash $test).Hash;fixture_sha256=(Get-FileHash $env:ASD_FIXTURE_TEST_BINARY).Hash;version=(& $env:ASD_PSMUX_TEST_BINARY -V|Out-String);candidate_processes_remaining=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $env:ASD_PSMUX_TEST_BINARY}).Count}
$r|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'receipt.json') -Encoding utf8;$r|ConvertTo-Json -Depth 5
if($code -ne 0 -or $r.candidate_processes_remaining){throw 'Native suite or cleanup failed'}
