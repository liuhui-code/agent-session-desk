$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'components-r3-worker.ps1') -Task 'AsdInstanceComponentsR320261126' -Receipt (Join-Path $base 'instance-components-r3-20261126/result.json') -Seconds 150
