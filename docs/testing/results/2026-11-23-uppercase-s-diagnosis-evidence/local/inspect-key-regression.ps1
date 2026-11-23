$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'sidebar-key-diagnosis-20261123'
$files=@{};foreach($name in @('regression-failure.json','key-s-result.json','caps-before.json','caps-off.json','caps-current.json','caps-final.json','caps-after-control.json')){if(Test-Path (Join-Path $root $name)){$files[$name]=[IO.File]::ReadAllText((Join-Path $root $name))|ConvertFrom-Json}}
$last=Get-ChildItem (Join-Path $root 'desktop/responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1
@{task_state=[string](Get-ScheduledTask 'AsdSidebarKeyDesktop20261123').State;files=$files;last_response=$last.Name;last=([IO.File]::ReadAllText($last.FullName)|ConvertFrom-Json)}|ConvertTo-Json -Depth 12
