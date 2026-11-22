$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$sharedRoot=Join-Path $base 'tab-resources-20261123';$root=Join-Path $sharedRoot 'run3';$repo=Join-Path $env:USERPROFILE 'asdr-1123';$name='AsdTabResources20261123Run3'
$oldTask=Get-ScheduledTask -TaskName 'AsdTabResources20261123Run2'
$oldState=[IO.File]::ReadAllText((Join-Path $sharedRoot 'run2/runner-state.json'))|ConvertFrom-Json
if($oldTask.State -eq 'Running' -or $oldState.status -eq 'running'){throw 'Previous experiment is not terminal'}
$paths=@([IO.File]::ReadAllLines((Join-Path $sharedRoot 'run2/evidence/samples.jsonl'))|ConvertFrom-Json|ForEach-Object {$_.snapshot.processes}|Where-Object {$_.executable_path -like '*resource-runtime-*'}|ForEach-Object {$_.executable_path}|Sort-Object -Unique)
$remaining=@(Get-CimInstance Win32_Process|Where-Object {$_.ExecutablePath -and $paths -contains $_.ExecutablePath}|Select-Object ProcessId,CreationDate,ExecutablePath)
if($remaining.Count -ne 0){throw 'Previous experiment still has exact owned image processes'}
New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null
@{previous_task_state=[string]$oldTask.State;previous_runner_status=$oldState.status;owned_image_paths_checked=$paths.Count;remaining_exact_owned_images=$remaining;utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'preflight-cleanup.json') -Encoding utf8
foreach($file in @('windows-resources.ps1','windows-resource-ui.ps1','test-windows-resource-ui.ps1','windows-fixture-journal.ps1','test-windows-fixture-journal.ps1')){[IO.File]::Copy((Join-Path $sharedRoot $file),(Join-Path $repo ('tools/ci/'+$file)),$true)}
[IO.File]::Copy((Join-Path $sharedRoot 'expected-run3-sources.json'),(Join-Path $root 'expected-sources.json'))
[IO.File]::Copy((Join-Path $sharedRoot 'run-resource-run3.ps1'),(Join-Path $root 'run-resource-run3.ps1'))
$expected=[IO.File]::ReadAllText((Join-Path $root 'expected-sources.json'))|ConvertFrom-Json;$actual=[ordered]@{}
foreach($entry in $expected.PSObject.Properties){$hash=(Get-FileHash (Join-Path $repo $entry.Name)).Hash.ToLowerInvariant();if($hash -ne $entry.Value){throw ('Source mismatch: '+$entry.Name)};$actual[$entry.Name]=$hash}
$actual|ConvertTo-Json -Depth 4|Set-Content (Join-Path $root 'source-before.json') -Encoding utf8
$commit=(& git -C $repo rev-parse HEAD|Out-String).Trim();if($commit -ne '921b67013ac779e5f9d847e796d32acfb22e264a'){throw 'Base checkout changed'}
@{repo=$repo;commit=$commit;overlay=@('tools/ci/windows-resources.ps1','tools/ci/windows-resource-ui.ps1','tools/ci/test-windows-resource-ui.ps1','tools/ci/windows-fixture-journal.ps1','tools/ci/test-windows-fixture-journal.ps1');source_count=$actual.Count;utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json -Depth 4|Set-Content (Join-Path $root 'source-receipt.json') -Encoding utf8
if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){throw 'Resource task name already exists'}
$pwsh=Join-Path $base 'pwsh7/pwsh.exe'
if((Get-FileHash $pwsh).Hash -ne 'F2833E07A5A7D59153FFC63DB83BD6B62F6C63FFD8B5B5099A2FEADF6B605336'){throw 'PowerShell binary differs'}
$a=New-ScheduledTaskAction -Execute $pwsh -Argument ('-NoLogo -NoProfile -WindowStyle Hidden -File "'+(Join-Path $root 'run-resource-run3.ps1')+'"')
$p=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask -TaskName $name -Action $a -Principal $p -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 25))|Out-Null
Start-ScheduledTask $name
@{task=$name;started_utc=[DateTime]::UtcNow.ToString('o');source_count=$actual.Count;preflight_remaining=$remaining}|ConvertTo-Json -Depth 4
