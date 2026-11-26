$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'update-daily-worker-r2.ps1') -Task 'AsdPetFixDailyR220261126' -Receipt (Join-Path $base 'pet-fix-daily-r2-20261126/result.json') -Seconds 240
