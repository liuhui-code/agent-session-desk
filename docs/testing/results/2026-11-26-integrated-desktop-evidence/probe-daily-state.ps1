$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$target=Join-Path $package 'asd-integrated-74a86959.exe'
$r=@{}
foreach($pair in @(@('current','data/state/metadata.sqlite3'),@('backup','data/rollback-integrated-20261126/metadata-v5.sqlite3'),@('retained','data/rollback-integrated-20261126/metadata-after-update-retained.sqlite3'))){
 $p=Join-Path $package $pair[1];$exists=Test-Path $p;$r[$pair[0]+'_exists']=$exists
 if($exists){$data=(& $target metadata check $p 2>$null|Out-String);$r[$pair[0]+'_check_exit']=$LASTEXITCODE;if($LASTEXITCODE -eq 0){$c=$data|ConvertFrom-Json;$r[$pair[0]+'_schema']=$c.schema_version;$r[$pair[0]+'_integrity']=$c.integrity}}
}
$r.audit_receipt_exists=Test-Path (Join-Path $base 'daily-integrated-final-audit.json')
$r.audit_task_state=[string](Get-ScheduledTask -TaskName 'AsdIntegratedDailyAudit_20261126' -ErrorAction SilentlyContinue).State
$r.audit_task_result=(Get-ScheduledTaskInfo -TaskName 'AsdIntegratedDailyAudit_20261126' -ErrorAction SilentlyContinue).LastTaskResult
$r|ConvertTo-Json
