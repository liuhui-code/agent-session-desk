$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$out="$b\detach-diagnosis";$s=Get-Content "$out\created.json" -Raw|ConvertFrom-Json
$common=@('--backend',"$b\backend\psmux.exe",'--runtime-dir',"$b\detach-runtime",'--namespace','asd-detach-20261122')
$candidate="$b\asd-detach-guidance.exe";$live=& $candidate @common list --json|ConvertFrom-Json
if($LASTEXITCODE-ne 0 -or @($live).Count-ne 1 -or $live.session_id-ne$s.session_id -or $live.run_id-ne$s.run_id){throw 'Diagnostic Run changed'}
$journal=@(Get-Content "$out\fixture.jsonl" -Encoding UTF8|ForEach-Object{$_|ConvertFrom-Json});$starts=@($journal|Where-Object kind -eq started)
if($starts.Count-ne 1){throw 'Fixture restarted'}
$p=Get-Process -Id $starts[0].pid
$hash=(Get-FileHash $candidate).Hash.ToLowerInvariant()
if($hash-ne'5085d052c5468ced425936409f935628d01b7a6c10e9fb3559869afc6a5088b9'){throw 'Candidate hash mismatch'}
$exeProcess=@(Get-CimInstance Win32_Process -Filter "Name='asd-detach-guidance.exe'"|Where-Object ExecutablePath -eq $candidate|Select-Object ProcessId,ParentProcessId,ExecutablePath)
@{utc=[DateTime]::UtcNow.ToString('o');created=$s;live=$live;fixture_started_once=$true;fixture_pid=$p.Id;fixture_start_time=$p.StartTime.ToUniversalTime().ToString('o');candidate_sha256=$hash;candidate_processes=$exeProcess;psmux_sha256=(Get-FileHash "$b\backend\psmux.exe").Hash.ToLowerInvariant();fixture_journal=$journal}|ConvertTo-Json -Depth 10|Set-Content "$out\final-check.json" -Encoding UTF8
Copy-Item "$b\detach-runtime\asd-detach-20261122\input_debug.log" "$out\psmux-input-debug.log"
Get-Content "$out\final-check.json" -Raw
