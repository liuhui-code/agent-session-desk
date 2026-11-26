$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$manifest=Get-Content -Raw (Join-Path $package 'alpha.json')|ConvertFrom-Json
$old=Join-Path $package $manifest.entrypoint;$new=Join-Path $base 'asd-integrated-r3.exe'
if((Get-FileHash $old).Hash.ToLowerInvariant() -cne '2aac7ca51c497a8943fa0b342d289a4f159cb33e5a022c0db4a97eedc292291f'){throw 'Daily entry changed; no trial'}
if((Get-FileHash $new).Hash.ToLowerInvariant() -cne '74a86959355ba37ea641d08c8664602fd8daa8e8313c1b621fffe4986fe31d99'){throw 'Candidate changed'}
$trial=Join-Path $base 'daily-migration-trial-r3-20261126'
if(Test-Path $trial){throw 'Fresh local-only trial directory required'}
New-Item -ItemType Directory $trial|Out-Null
$common=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',(Join-Path $package 'data/runtime'),'--namespace','asd-alpha-8234c4e')
$sourceState=Join-Path $package 'data/state';$source=Join-Path $sourceState 'metadata.sqlite3'
& $new @common '--state-dir' $sourceState metadata backup (Join-Path $trial 'metadata.sqlite3') 1>$null 2>$null
if($LASTEXITCODE -ne 0){throw 'Validated local snapshot failed'}
function HistoryStamp([string]$exe){
 $rows=(& $exe @common '--state-dir' $trial history --limit 100 2>$null|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0){throw 'Local-only history read failed'}
 $selected=@($rows|ForEach-Object{$s=$_.session;[ordered]@{session=$s.session_id;run=$s.run_id;backend=$s.backend_id;title=$s.title;title_revision=$s.title_revision;command=$s.command;group=$s.group_id;group_revision=$s.group_revision;position=$s.order;position_revision=$s.order_revision}}|Sort-Object {$_.session},{$_.run})
 $bytes=[Text.Encoding]::UTF8.GetBytes(($selected|ConvertTo-Json -Depth 5 -Compress))
 $sha=[Security.Cryptography.SHA256]::Create();try{$stamp=([BitConverter]::ToString($sha.ComputeHash($bytes))).Replace('-','')}finally{$sha.Dispose()}
 return @{stamp=$stamp;count=$selected.Count}
}
$before=HistoryStamp $old
& $new @common '--state-dir' $trial attention-list 1>$null 2>$null
if($LASTEXITCODE -ne 0){throw 'Clone migration failed'}
$after=HistoryStamp $new
$sourceCheck=(& $new metadata check $source 2>$null|Out-String)|ConvertFrom-Json
if($LASTEXITCODE -ne 0){throw 'Original metadata validation failed'}
$cloneCheck=(& $new metadata check (Join-Path $trial 'metadata.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
if($LASTEXITCODE -ne 0){throw 'Clone metadata validation failed'}
$result=@{source_schema=[int]$sourceCheck.schema_version;clone_schema=[int]$cloneCheck.schema_version;source_integrity_ok=($sourceCheck.integrity -ceq 'ok');clone_integrity_ok=($cloneCheck.integrity -ceq 'ok');bounded_history_records=$before.count;history_projection_preserved=($before.stamp -ceq $after.stamp);source_database_migrated=$false;session_rows_exported=0;private_snapshot_remains_on_windows=$true;output_records_not_copied=$true}
$result|ConvertTo-Json|Set-Content (Join-Path $base 'daily-migration-trial-summary.json') -Encoding utf8
$result|ConvertTo-Json
if($result.source_schema -ne 5 -or $result.clone_schema -ne 9 -or !$result.history_projection_preserved){throw 'Migration trial not accepted; daily state unchanged'}
