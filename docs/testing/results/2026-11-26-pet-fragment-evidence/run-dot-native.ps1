$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'dot-unit-worker.ps1') -Task 'AsdPetDotUnits20261126' -Receipt (Join-Path $base 'pet-dot-native-20261126/result.json') -Seconds 60
