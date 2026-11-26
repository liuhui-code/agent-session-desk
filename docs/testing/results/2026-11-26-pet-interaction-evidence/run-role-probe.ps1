$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'probe-roles-worker.ps1') -Task 'AsdPetFixRoles20261126' -Receipt (Join-Path $base 'pet-fix-role-probe-20261126.json') -Seconds 60
