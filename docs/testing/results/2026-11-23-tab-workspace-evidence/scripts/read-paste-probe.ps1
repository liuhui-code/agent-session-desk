$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/desktop'
if(Test-Path (Join-Path $root 'paste-probe.json')){Get-Content (Join-Path $root 'paste-probe.json') -Raw -Encoding utf8}
Get-ScheduledTaskInfo -TaskName AsdTabsPasteProbe20261123|Select-Object LastTaskResult|ConvertTo-Json
