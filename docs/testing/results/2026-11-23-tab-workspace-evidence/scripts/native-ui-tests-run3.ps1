$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'tabs-ui-20261123'
if(!(Test-Path $root)){throw 'Native run root missing'}
@{utc=[DateTime]::UtcNow.ToString('o');exe=(Get-FileHash (Join-Path $base 'tabs-ui-tests-run3.exe')).Hash;app=(Get-FileHash (Join-Path $base 'tabs-ui-run3.exe')).Hash;windows=[Environment]::OSVersion.VersionString}|ConvertTo-Json|Set-Content (Join-Path $root 'binaries-run3.json') -Encoding utf8
& (Join-Path $base 'tabs-ui-tests-run3.exe') '-test.v' *> (Join-Path $root 'go-ui-run3.log')
$code=$LASTEXITCODE
Get-Content (Join-Path $root 'go-ui-run3.log') -Tail 8
if($code -ne 0){throw 'Native UI tests failed'}
