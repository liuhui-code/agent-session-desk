$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'tabs-ui-20261123/native-backend-run3'
if(Test-Path $root){throw 'Fresh evidence directory required'}
New-Item -ItemType Directory $root|Out-Null
$env:ASD_PSMUX_TEST_BINARY=Join-Path $base 'backend/psmux.exe'
$env:ASD_PSMUX_TEST_SHA256='bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'
$env:ASD_FIXTURE_TEST_BINARY=Join-Path $base 'asd-fixture.exe'
$env:ASD_NATIVE_EVIDENCE_DIR=Join-Path $root 'native'
$env:ASD_SHELL_TEST_BINARY=Join-Path $base 'pwsh7/pwsh.exe'
$test=Join-Path $base 'tabs-backend-tests-run3.exe'
$hash=(Get-FileHash $test).Hash.ToLowerInvariant()
& $test '-test.run=^TestWindowsPsmuxNativeTerminal' '-test.v' '-test.timeout=3m' *> (Join-Path $root 'tests.log')
$code=$LASTEXITCODE
Get-Content (Join-Path $root 'tests.log')
@{exit_code=$code;test_sha256=$hash;utc=[DateTime]::UtcNow.ToString('o');windows_build=[Environment]::OSVersion.Version.ToString()}|ConvertTo-Json|Set-Content (Join-Path $root 'result.json') -Encoding utf8
if($code -ne 0){throw "Native tests exited $code"}
