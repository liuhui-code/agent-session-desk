$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'unit-worker.ps1') -Task 'AsdPetFixUnits20261126' -Receipt (Join-Path $base 'pet-fix-native-20261126/tests.json') -Seconds 200
