$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-rename-desktop-20261123';$driver=Join-Path $root 'desktop'
$files=@(Get-ChildItem (Join-Path $driver 'responses') -File|Sort-Object Name)
$actions=@();foreach($file in $files){$r=[IO.File]::ReadAllText($file.FullName)|ConvertFrom-Json;$actions+=,@{id=$file.Name;ok=$r.ok;kind=$r.command.kind;text=$r.command.text;keys=$r.command.keys}}
$last=($files|Select-Object -Last 1);$state=[IO.File]::ReadAllText($last.FullName)|ConvertFrom-Json
@{task=[string](Get-ScheduledTask -TaskName AsdSessionRenameDesktop20261123).State;pending=@(Get-ChildItem (Join-Path $driver 'commands') -File).Count;done=@(Get-ChildItem (Join-Path $driver 'done') -File).Count;actions=$actions;last_state=$state.state.text;failure=([IO.File]::ReadAllText((Join-Path $root 'rename-resume-failure.json'))|ConvertFrom-Json)}|ConvertTo-Json -Depth 7
