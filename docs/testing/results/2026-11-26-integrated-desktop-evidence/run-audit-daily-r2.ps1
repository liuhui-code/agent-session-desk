$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'audit-daily-worker-r2.ps1') -Task 'AsdIntegratedDailyAuditR2_20261126' -Receipt (Join-Path $base 'daily-integrated-final-audit-r2.json') -Seconds 60
