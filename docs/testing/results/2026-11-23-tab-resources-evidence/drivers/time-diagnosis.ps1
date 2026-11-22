$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tab-resources-20261123/run2'
$inner=@'
$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tab-resources-20261123/run2'
$summary=[IO.File]::ReadAllText((Join-Path $root 'evidence/summary.json'))|ConvertFrom-Json
$journal=[IO.File]::ReadAllLines((Join-Path $root 'evidence/managed-1-fixture-0.jsonl'))|ConvertFrom-Json
$done=@($journal|Where-Object event -eq 'load-finished')[0]
$samples=[IO.File]::ReadAllLines((Join-Path $root 'evidence/samples.jsonl'))|ConvertFrom-Json
$last=@($samples|Where-Object {$_.scenario -eq 'managed-1' -and $_.phase -eq 'selected-output'})[-1]
$events=[IO.File]::ReadAllLines((Join-Path $root 'evidence/events.jsonl'))|ConvertFrom-Json
$ownedPaths=@($events|Where-Object event -eq 'forced-owned-cleanup'|ForEach-Object {$_.data.identity.executable_path}|Sort-Object -Unique)
$remaining=@(Get-CimInstance Win32_Process|Where-Object {$_.ExecutablePath -and $ownedPaths -contains $_.ExecutablePath}|Select-Object ProcessId,CreationDate,ExecutablePath)
@{powershell=$PSVersionTable.PSVersion.ToString();timezone=[TimeZoneInfo]::Local.Id;current_culture=[Globalization.CultureInfo]::CurrentCulture.Name;typed_finish=$done.finished_utc.ToString('o');finish_type=$done.finished_utc.GetType().FullName;finish_kind=[string]$done.finished_utc.Kind;reparsed_finish=[DateTime]::Parse($done.finished_utc).ToUniversalTime().ToString('o');direct_finish=$done.finished_utc.ToUniversalTime().ToString('o');last_sample=$last;selected_output_phase=@($summary.phases|Where-Object phase -eq 'selected-output');remaining_exact_owned_images=$remaining;forced_cleanup_count=@($events|Where-Object event -eq 'forced-owned-cleanup').Count}|ConvertTo-Json -Depth 8
'@
$path=Join-Path $root 'time-diagnosis-inner.ps1';$inner|Set-Content $path -Encoding utf8
& (Join-Path $base 'pwsh7/pwsh.exe') -NoLogo -NoProfile -File $path *> (Join-Path $root 'time-diagnosis.json')
if($LASTEXITCODE -ne 0){throw 'Timestamp diagnosis failed'}
[IO.File]::ReadAllText((Join-Path $root 'time-diagnosis.json'))
