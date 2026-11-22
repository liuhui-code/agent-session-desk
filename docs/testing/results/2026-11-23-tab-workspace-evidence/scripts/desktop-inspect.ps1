$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tabs-ui-20261123/desktop'
Get-ScheduledTask -TaskName AsdTabsDesktop20261123|Select-Object TaskName,State|ConvertTo-Json
Get-ScheduledTaskInfo -TaskName AsdTabsDesktop20261123|Select-Object LastRunTime,LastTaskResult|ConvertTo-Json
if(Test-Path $root){Get-ChildItem $root|Select-Object Name,Length|ConvertTo-Json;foreach($name in @('ready.json','driver-error.json')){if(Test-Path (Join-Path $root $name)){Get-Content (Join-Path $root $name) -Raw -Encoding utf8}}}
