$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new();$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$asd="$b\product\asd-windows-amd64.exe";$o=@('--backend',"$b\backend\psmux.exe",'--runtime-dir',"$b\desktop-runtime",'--namespace','asd-desktop-20261122')
$pair=Get-Content "$b\desktop\second-pair.json" -Raw|ConvertFrom-Json
function Flat($x){if($x-is[Array]){foreach($v in $x){Flat $v}}elseif($x.PSObject.Properties['session_id']){$x}elseif($x.PSObject.Properties['value']){Flat $x.value}}
$pair=@(Flat $pair);$live=@(Flat (& $asd @o list --json|ConvertFrom-Json));$logs=@{};$processes=@();$stops=@()
foreach($n in @('A2','B2')){$logs[$n]=@(Get-Content "$b\desktop\fixture-$n.jsonl" -Encoding UTF8|ForEach-Object{$_|ConvertFrom-Json});$start=$logs[$n]|Where-Object kind -eq started;$p=Get-Process -Id $start.pid;$processes+=@{fixture=$n;pid=$p.Id;start_time=$p.StartTime.ToUniversalTime().ToString('o')}}
foreach($c in $pair){if(@($live|Where-Object {$_.session_id-eq$c.session_id-and$_.run_id-eq$c.run_id}).Count-ne 1){throw 'Pair identity mismatch'}}
if(@($logs.A2|Where-Object kind -eq input).Count-ne 1 -or ($logs.A2|Where-Object kind -eq input).line-cne'ASD_RETURN_A_FIRST'){throw 'A2 record mismatch'}
if(@($logs.B2|Where-Object kind -eq input).Count-ne 1 -or ($logs.B2|Where-Object kind -eq input).line-cne'ASD_RETURN_B_ONLY'){throw 'B2 record mismatch'}
foreach($c in $pair){$stops+=,(& $asd @o stop $c.session_id|ConvertFrom-Json);if($LASTEXITCODE-ne 0){throw 'Pair stop failed'}}
Start-Sleep -Milliseconds 400
$after=& $asd @o list --json
foreach($x in $processes){$p=Get-Process -Id $x.pid -ErrorAction SilentlyContinue;if($p-and$p.StartTime.ToUniversalTime().ToString('o')-eq$x.start_time){throw 'Original pair process still alive'}}
@{utc=[DateTime]::UtcNow.ToString('o');pair=$pair;live_before=$live;journals=$logs;processes=$processes;stop_results=$stops;after_list_raw=($after-join"`n");identity_preserved=$true;input_isolation=$true;original_processes_exited=$true}|ConvertTo-Json -Depth 12|Set-Content "$b\desktop\pair-final-check.json" -Encoding UTF8
Get-Content "$b\desktop\pair-final-check.json" -Raw
