$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tab-resources-20261123';$task=Get-ScheduledTask -TaskName 'AsdTabResources20261123Run3';$last=[IO.File]::ReadAllText((Join-Path $root 'run3/runner-state.json'))|ConvertFrom-Json
if($task.State -eq 'Running' -or $last.status -eq 'running'){throw 'Run3 not terminal; evidence export did not start'}
$export=Join-Path $root 'evidence-export';New-Item -ItemType Directory -Path $export -ErrorAction Stop|Out-Null
$run1=Join-Path $export 'run1';New-Item -ItemType Directory -Path $run1|Out-Null
foreach($file in @('run-resource.ps1','runner-state.json','runner-error.txt','source-receipt.json','expected-sources.json','source-before.json','source-after.json','native-ui-helper-tests.log','native-process-metrics-tests.log')){if(Test-Path (Join-Path $root $file)){[IO.File]::Copy((Join-Path $root $file),(Join-Path $run1 $file))}}
foreach($dir in @('run2','run3','focus-probe','focus-key-probe','focus-key-probe2','focus-text-probe','focus-hex-probe','focus-protocol-probe','focus-uv-probe')){Copy-Item -LiteralPath (Join-Path $root $dir) -Destination (Join-Path $export $dir) -Recurse}
foreach($file in @('red-01-bootstrap.log','red-01.log','green-01.log','red-02.log','green-02.log','time-red-03.log','time-green-03.log','build-receipt.json','product-build-info.log','go-version.log','resource-focus-probe.ps1','resource-focus-text-probe.ps1','resource-focus-hex-probe.ps1','resource-focus-protocol-probe.ps1','focus-probe.log','focus-key-probe.log','focus-key-probe2.log','focus-text-probe.log','focus-hex-probe.log','focus-protocol-probe.log','focus-uv-probe.log','focus-key-events.log','focus-key-events2.log','focus-text-key-events.log','focus-hex-key-events.log','focus-protocol-key-events.log','focus-uv-key-events.log')){if(Test-Path (Join-Path $root $file)){[IO.File]::Copy((Join-Path $root $file),(Join-Path $export $file))}}
$allPaths=@()
foreach($dir in @('run2/evidence','run3/evidence','focus-probe','focus-key-probe','focus-key-probe2','focus-text-probe','focus-hex-probe','focus-protocol-probe','focus-uv-probe')){
 $file=Join-Path $root ($dir+'/samples.jsonl')
 if(Test-Path $file){$allPaths+=@([IO.File]::ReadAllLines($file)|ConvertFrom-Json|ForEach-Object {$_.snapshot.processes}|Where-Object {$_.executable_path -like '*resource-runtime-*'}|ForEach-Object {$_.executable_path})}
}
$paths=@($allPaths|Sort-Object -Unique)
$remaining=@(Get-CimInstance Win32_Process|Where-Object {$_.ExecutablePath -and $paths -contains $_.ExecutablePath}|Select-Object ProcessId,CreationDate,ExecutablePath)
$tasks=@('AsdTabResources20261123','AsdTabResources20261123Run2','AsdTabResources20261123Run3')
$taskResults=@(foreach($name in $tasks){$t=Get-ScheduledTask -TaskName $name;$i=Get-ScheduledTaskInfo -TaskName $name;[pscustomobject]@{name=$name;state=[string]$t.State;last_result=$i.LastTaskResult}})
if(@($taskResults|Where-Object state -eq 'Running').Count -gt 0 -or $remaining.Count -gt 0){throw 'A known experiment remains running; tasks were not removed'}
foreach($name in $tasks){Unregister-ScheduledTask -TaskName $name -Confirm:$false}
@{task_results_before_removal=$taskResults;remaining_exact_owned_images=$remaining;exact_image_paths_checked=$paths.Count;remaining_tasks=@(Get-ScheduledTask|Where-Object {$tasks -contains $_.TaskName}|Select-Object TaskName,State);utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $export 'cleanup.json') -Encoding utf8
$zip=Join-Path $root 'evidence-export.zip';Compress-Archive -Path (Join-Path $export '*') -DestinationPath $zip -ErrorAction Stop
$summary=[IO.File]::ReadAllText((Join-Path $root 'run3/evidence/summary.json'))|ConvertFrom-Json
@{runner_status=$last.status;resource_result=$summary.result;failure=$summary.failure;scenarios=$summary.scenarios.Count;phases=$summary.phases.Count;samples=$summary.sample_count;normal_cleanups=@($summary.cleanup|Where-Object normal_cleanup -eq $true).Count;elapsed_ms=$summary.elapsed_ms;export_bytes=(Get-Item $zip).Length;export_sha256=(Get-FileHash $zip).Hash}|ConvertTo-Json
