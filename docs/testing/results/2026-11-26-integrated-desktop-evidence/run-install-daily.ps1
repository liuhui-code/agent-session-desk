$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'install-daily-worker.ps1') -Task 'AsdIntegratedDailyInstall_20261126' -Receipt (Join-Path $base 'daily-integrated-install-20261126/result.json') -Seconds 240
