$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'audit-daily-worker.ps1') -Task 'AsdIntegratedDailyAudit_20261126' -Receipt (Join-Path $base 'daily-integrated-final-audit.json') -Seconds 60
