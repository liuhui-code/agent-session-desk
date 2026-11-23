$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'history-desktop-20261123';$d=Join-Path $root 'desktop'
$task=Get-ScheduledTask AsdHistoryDesktop20261123 -ErrorAction Stop;$failure=[IO.File]::ReadAllText((Join-Path $root 'flow-failure.json'))|ConvertFrom-Json
$pending=@(Get-ChildItem (Join-Path $d 'commands') -Filter '*.json');if($pending.Count){throw 'Pending driver action; do not enqueue another'}
$done=@(Get-ChildItem (Join-Path $d 'done') -Filter '*.json'|Sort-Object Name);$n=if($done.Count){[int]$done[-1].BaseName+1}else{1}
$path=Join-Path (Join-Path $d 'commands') ('{0:d3}.json'-f$n);[IO.File]::WriteAllText($path,'{"kind":"state"}',[Text.UTF8Encoding]::new($false))
$response=Join-Path (Join-Path $d 'responses') ('{0:d3}.json'-f$n);$until=[DateTime]::UtcNow.AddSeconds(10);while(!(Test-Path $response)-and[DateTime]::UtcNow-lt$until){Start-Sleep -Milliseconds 100}
$state=[IO.File]::ReadAllText($response)|ConvertFrom-Json
$identities=@();foreach($name in@('Copy_A','Copy_B')){$i=[IO.File]::ReadAllText((Join-Path $root ('work/'+$name+'.identity.json')))|ConvertFrom-Json;$p=Get-Process -Id $i.pid;$identities+=,@{pid=$i.pid;filetime=$i.filetime;path=$i.path;same=($p.Path-eq$i.path-and$p.StartTime.ToUniversalTime().ToFileTimeUtc()-eq$i.filetime)}}
$r=@{task_state=[string]$task.State;failed_sequence=$failure.sequence;failure=$failure.error;last_sequence=$n;state_ok=$state.ok;app_text=$state.state.text;shells=$identities}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'failure-inspection.json') -Encoding utf8
@{task_state=$r.task_state;failed_sequence=$r.failed_sequence;last_sequence=$n;state_ok=$state.ok;app_text=$r.app_text;shells_same=@($identities|Where-Object{$_.same}).Count}|ConvertTo-Json -Depth 5
