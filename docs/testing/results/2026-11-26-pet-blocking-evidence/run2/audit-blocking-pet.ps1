$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$names=@('asd-blocking-pet.exe','asd-fixture-blocking-pet.exe','pet-blocking.test.exe','ui-pet-blocking.test.exe','cli-pet-blocking.test.exe')
$matches=@(Get-CimInstance Win32_Process|Where-Object{$_.Name -in $names -and $_.ExecutablePath -ieq (Join-Path $base $_.Name)}|ForEach-Object{@{pid=$_.ProcessId;name=$_.Name;path=$_.ExecutablePath;created=$_.CreationDate.ToUniversalTime().ToString('o')}})
$tasks=@(foreach($name in @('AsdBlockingPet20261126Run1','AsdBlockingPet20261126Run2')){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){$name}})
$result=@{owned_images=$matches;temporary_tasks=$tasks;image_audit_count=$matches.Count;task_audit_count=$tasks.Count;daily_entry_modified=$false;personal_state_migrated=$false;forced_ends=0;clipboard_calls=0;backend_global_audit=$false}
$result|ConvertTo-Json -Depth 5|Set-Content (Join-Path $base 'blocking-pet-audit.json') -Encoding utf8
$result|ConvertTo-Json -Depth 5
if($matches.Count -gt 0 -or $tasks.Count -gt 0){throw 'Owned verification resources remain; no broad cleanup attempted'}
