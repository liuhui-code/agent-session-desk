$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$tests=Get-Content -Raw (Join-Path $base 'pet-fix-native-20261126/tests.json')|ConvertFrom-Json
if($tests.Count -ne 2 -or @($tests|Where-Object{$_.exit_code -ne 0 -or $_.top_fail -ne 0 -or $_.top_skip -ne 0}).Count){throw 'Native units not accepted'}
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'desktop-pet-fix.ps1') -Task 'AsdPetFixDesktop20261126' -Receipt (Join-Path $base 'pet-fix-desktop-20261126/result.json') -Seconds 240
