$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tab-resources-20261123';$repo=Join-Path $env:USERPROFILE 'asdr-1123'
$last=[IO.File]::ReadAllText((Join-Path $root 'run3/runner-state.json'))|ConvertFrom-Json
if((Get-ScheduledTask -TaskName 'AsdTabResources20261123Run3').State -eq 'Running' -or $last.status -eq 'running'){throw 'Run3 still running'}
$paths=@([IO.File]::ReadAllLines((Join-Path $root 'run3/evidence/samples.jsonl'))|ConvertFrom-Json|ForEach-Object {$_.snapshot.processes}|Where-Object {$_.executable_path -like '*resource-runtime-*'}|ForEach-Object {$_.executable_path}|Sort-Object -Unique)
if(@(Get-CimInstance Win32_Process|Where-Object {$_.ExecutablePath -and $paths -contains $_.ExecutablePath}).Count -ne 0){throw 'Previous owned images remain'}
[IO.File]::Copy((Join-Path $root 'resource-focus-hex-probe.ps1'),(Join-Path $repo 'tools/ci/resource-focus-hex-probe.ps1'),$true)
$env:ASD_DIAGNOSTIC_KEYS=Join-Path $root 'focus-hex-key-events.log'
$ErrorActionPreference='Continue'
& (Join-Path $base 'pwsh7/pwsh.exe') -NoLogo -NoProfile -File (Join-Path $repo 'tools/ci/resource-focus-hex-probe.ps1') -PsmuxBinary (Join-Path $base 'backend/psmux.exe') -ProductBinary (Join-Path $root 'asd-key-probe2.exe') -FixtureBinary (Join-Path $root 'asd-fixture-resource.exe') -EvidenceDirectory (Join-Path $root 'focus-hex-probe') *> (Join-Path $root 'focus-hex-probe.log')
$code=$LASTEXITCODE;$ErrorActionPreference='Stop'
@{probe_exit_code=$code;summary=([IO.File]::ReadAllText((Join-Path $root 'focus-hex-probe/summary.json'))|ConvertFrom-Json);probes=@([IO.File]::ReadAllLines((Join-Path $root 'focus-hex-probe/events.jsonl'))|ConvertFrom-Json|Where-Object event -like 'probe-*'|ForEach-Object {[pscustomobject]@{event=$_.event;focus=$_.data.focus}})}|ConvertTo-Json -Depth 4
if(Test-Path (Join-Path $root 'focus-hex-key-events.log')){[IO.File]::ReadAllText((Join-Path $root 'focus-hex-key-events.log'))}
if($code -ne 0){throw 'Focus diagnostic failed'}
