$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'f6-reader-20261123';$run=Join-Path $root 'resource-sdk1';$name='AsdTabResources20261123Sdk1'
$task=Get-ScheduledTask -TaskName $name;$info=Get-ScheduledTaskInfo -TaskName $name
$state=[IO.File]::ReadAllText((Join-Path $run 'runner-state.json'))|ConvertFrom-Json
if($task.State -eq 'Running' -or $state.status -eq 'running'){throw 'Resource task is not terminal; export did not start'}
$known=@()
foreach($part in @('evidence','candidate-evidence','candidate2-evidence','resource-sdk1/evidence')){
 $known+=@([IO.File]::ReadAllLines((Join-Path $root ($part+'/samples.jsonl')))|ConvertFrom-Json|ForEach-Object {$_.snapshot.processes})
}
$live=@()
foreach($identity in @($known|Sort-Object process_id,creation_filetime -Unique)){
 if($identity.executable_path -notlike '*resource-runtime-*' -and $identity.executable_path -notlike '*f6-reader-20261123*candidate-native*'){continue}
 $process=Get-Process -Id $identity.process_id -ErrorAction SilentlyContinue
 if($process){try{if($process.StartTime.ToUniversalTime().ToFileTimeUtc().ToString() -eq $identity.creation_filetime -and $process.Path -eq $identity.executable_path){$live+=$identity}}finally{$process.Dispose()}}
}
if($live.Count -ne 0){throw 'Exact owned processes remain; task was not removed'}
$export=Join-Path $root 'evidence-export';New-Item -ItemType Directory $export -ErrorAction Stop|Out-Null
foreach($dir in @('evidence','candidate-evidence','candidate2-evidence','resource-sdk1')){Copy-Item -LiteralPath (Join-Path $root $dir) -Destination (Join-Path $export $dir) -Recurse}
foreach($file in @('probe.log','candidate-probe.log','candidate2-probe.log','candidate-verification.json','driver-red01.log','driver-green01.log','driver-red02.log','driver-green02.log')){[IO.File]::Copy((Join-Path $root $file),(Join-Path $export $file))}
Unregister-ScheduledTask -TaskName $name -Confirm:$false
@{task_state_before_removal=[string]$task.State;task_result=$info.LastTaskResult;runner_state=$state.status;exact_identities_checked=@($known|Sort-Object process_id,creation_filetime -Unique).Count;remaining_exact_owned=$live;remaining_task=@(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue|Select-Object TaskName,State);utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $export 'cleanup.json') -Encoding utf8
$zip=Join-Path $root 'evidence-export.zip';Compress-Archive -Path (Join-Path $export '*') -DestinationPath $zip -ErrorAction Stop
$summary=[IO.File]::ReadAllText((Join-Path $run 'evidence/summary.json'))|ConvertFrom-Json
@{result=$summary.result;failure=$summary.failure;runner_status=$state.status;scenarios=$summary.scenarios.Count;phases=$summary.phases.Count;samples=$summary.sample_count;normal_cleanups=@($summary.cleanup|Where-Object normal_cleanup -eq $true).Count;elapsed_ms=$summary.elapsed_ms;export_sha256=(Get-FileHash $zip).Hash;export_bytes=(Get-Item $zip).Length}|ConvertTo-Json
