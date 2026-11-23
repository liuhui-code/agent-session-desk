$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-rename-desktop-20261123';$driver=Join-Path $root 'desktop'
$files=@(Get-ChildItem (Join-Path $driver 'responses') -File|Sort-Object Name)
$responses=@();foreach($f in @($files|Select-Object -Last 6)){$r=[IO.File]::ReadAllText($f.FullName)|ConvertFrom-Json;$responses+=,@{id=$f.Name;ok=$r.ok;kind=$r.command.kind;keys=$r.command.keys;text=$r.command.text;state=$r.state.text;error=$r.error}}
$pending=@();foreach($f in Get-ChildItem (Join-Path $driver 'commands') -File){$pending+=,@{id=$f.Name;command=([IO.File]::ReadAllText($f.FullName)|ConvertFrom-Json)}}
$errorFile=Join-Path $driver 'driver-error.json';$driverError=$null;if(Test-Path $errorFile){$driverError=[IO.File]::ReadAllText($errorFile)|ConvertFrom-Json}
@{task_state=[string](Get-ScheduledTask -TaskName AsdSessionRenameDesktop20261123).State;pending=$pending;done=@(Get-ChildItem (Join-Path $driver 'done') -File).Count;responses=$responses;driver_error=$driverError;failure=([IO.File]::ReadAllText((Join-Path $root 'rename-final-failure.json'))|ConvertFrom-Json)}|ConvertTo-Json -Depth 9
