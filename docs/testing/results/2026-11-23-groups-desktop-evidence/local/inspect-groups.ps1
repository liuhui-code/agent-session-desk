$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'group-workspace-desktop-20261123';$desktop=Join-Path $root 'desktop'
function Shared($path){$f=[IO.File]::Open($path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete));try{$r=[IO.StreamReader]::new($f,[Text.Encoding]::UTF8,$true);try{return $r.ReadToEnd()}finally{$r.Dispose()}}finally{$f.Dispose()}}
$task=Get-ScheduledTask 'AsdGroupWorkspaceDesktop20261123' -ErrorAction SilentlyContinue
$r=@{task_state=[string]$task.State;utc=[DateTime]::UtcNow.ToString('o')}
if(Test-Path (Join-Path $desktop 'ready.json')){$ready=(Shared (Join-Path $desktop 'ready.json'))|ConvertFrom-Json;$p=Get-Process -Id $ready.pid -ErrorAction SilentlyContinue;$r.ready_identity_live=[bool]($p -and $p.Path -ceq $ready.path -and $p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $ready.filetime);$r.ready_pid=$ready.pid}
$files=@(Get-ChildItem (Join-Path $desktop 'responses') -Filter '*.json' -ErrorAction SilentlyContinue|Sort-Object Name)
if($files.Count){$last=(Shared $files[-1].FullName)|ConvertFrom-Json;$r.last_receipt=$files[-1].Name;$r.last_kind=$last.command.kind;$r.last_ok=$last.ok;if($last.error){$r.last_error=$last.error};if($last.state){$r.last_text=$last.state.text.Replace($env:USERPROFILE,'<WINDOWS_TEST_HOME>')}}
foreach($name in @('flow-result.json','flow-failure.json','desktop/driver-error.json')){if(Test-Path (Join-Path $root $name)){$r[$name]=((Shared (Join-Path $root $name))|ConvertFrom-Json)}}
$r|ConvertTo-Json -Depth 6
