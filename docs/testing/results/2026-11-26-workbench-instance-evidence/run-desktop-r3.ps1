$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$tests=Get-Content -Raw (Join-Path $base 'instance-components-r3-20261126/result.json')|ConvertFrom-Json
if($tests.Count-ne 3-or @($tests|Where-Object{$_.exit_code-ne 0-or $_.top_fail-ne 0-or $_.top_skip-ne 0}).Count){throw 'Native components not accepted; no desktop launch'}
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'instance-desktop-r3.ps1') -Task 'AsdInstanceDesktopR320261126' -Receipt (Join-Path $base 'instance-desktop-r3-20261126/result.json') -Seconds 270
