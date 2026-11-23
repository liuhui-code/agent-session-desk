$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'bundle-desktop-20261123'
$responses=@(Get-ChildItem (Join-Path $root 'desktop/responses') -Filter '*.json'|Sort-Object Name)
$last=[IO.File]::ReadAllText($responses[-1].FullName)|ConvertFrom-Json
$ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json
@{last_sequence=$responses[-1].BaseName;last=$last;startup=$ready;task_state=[string](Get-ScheduledTask AsdBundleDesktop20261123).State;product_processes=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq (Join-Path $base 'bundle-product.exe')}|ForEach-Object{@{pid=$_.ProcessId;path=$_.ExecutablePath}})}|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'startup-inspection.json') -Encoding utf8
@{sequence=$responses[-1].BaseName;state=$last.state.text;startup=$ready.text;task=[string](Get-ScheduledTask AsdBundleDesktop20261123).State}|ConvertTo-Json -Depth 4
