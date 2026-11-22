$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'tabs-ui-20261123'
New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null
@{utc=[DateTime]::UtcNow.ToString('o');exe=(Get-FileHash (Join-Path $base 'tabs-ui-tests.exe')).Hash;app=(Get-FileHash (Join-Path $base 'tabs-ui.exe')).Hash;windows=[Environment]::OSVersion.VersionString}|ConvertTo-Json|Set-Content (Join-Path $root 'binaries.json') -Encoding utf8
& (Join-Path $base 'tabs-ui-tests.exe') -test.v *> (Join-Path $root 'go-ui.log')
$code=$LASTEXITCODE
Get-Content (Join-Path $root 'go-ui.log') -Tail 8
if($code -ne 0){throw 'Native UI tests failed'}
