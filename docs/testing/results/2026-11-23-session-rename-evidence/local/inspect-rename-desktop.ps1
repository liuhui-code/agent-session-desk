$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-rename-desktop-20261123';$driver=Join-Path $root 'desktop'
$task=Get-ScheduledTask -TaskName AsdSessionRenameDesktop20261123
$responses=@(Get-ChildItem (Join-Path $driver 'responses') -File|Sort-Object Name)
$records=@();foreach($file in $responses){$r=[IO.File]::ReadAllText($file.FullName)|ConvertFrom-Json;$records+=,@{id=$file.Name;ok=$r.ok;kind=$r.command.kind;input_text=$r.command.text;input_keys=$r.command.keys;state_text=$r.state.text}}
$ready=[IO.File]::ReadAllText((Join-Path $driver 'ready.json'))|ConvertFrom-Json;$p=Get-Process -Id $ready.pid
if($p.Path -ne $ready.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Original window changed'}
$ids=@();foreach($name in @('Copy_A','Copy_B','中文会话')){$id=[IO.File]::ReadAllText((Join-Path $root ('work/'+$name+'.identity.json')))|ConvertFrom-Json;$q=Get-Process -Id $id.pid;if($q.Path -ne $id.path -or $q.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime){throw 'Original shell changed'};$ids+=,$id}
@{task_state=[string]$task.State;pending=@(Get-ChildItem (Join-Path $driver 'commands') -File).Count;done=@(Get-ChildItem (Join-Path $driver 'done') -File).Count;responses=$records;shells=$ids;failure=([IO.File]::ReadAllText((Join-Path $root 'rename-failure.json'))|ConvertFrom-Json)}|ConvertTo-Json -Depth 7
