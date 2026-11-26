$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$target=Join-Path $package 'asd-integrated-74a86959.exe';$old=Join-Path $package 'asd-focus-3654c0e.exe'
$compare=Join-Path $base 'daily-final-private-comparison-20261126'
$g=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',(Join-Path $package 'data/runtime'),'--namespace','asd-alpha-8234c4e')
if(!(Test-Path (Join-Path $compare 'metadata.sqlite3'))){throw 'Validated comparison clone required'}
$retained=Join-Path $package 'data/rollback-integrated-20261126/metadata-after-update-retained.sqlite3'
$restored=Join-Path $base 'daily-retained-private-compare-20261126'
if(Test-Path $restored){throw 'Fresh retained comparison required'}
New-Item -ItemType Directory $restored|Out-Null
& $target @g '--state-dir' $restored metadata restore $retained 1>$null 2>$null
if($LASTEXITCODE -ne 0){throw 'Retained comparison snapshot failed'}
function Projection($rows){$selected=@($rows|ForEach-Object{$s=$_.session;[ordered]@{session=$s.session_id;run=$s.run_id;backend=$s.backend_id;command=$s.command;title=$s.title;title_revision=$s.title_revision;group=$s.group_id;group_revision=$s.group_revision;position=$s.order;position_revision=$s.order_revision}}|Sort-Object {$_.session},{$_.run});return ConvertTo-Json -InputObject $selected -Depth 6 -Compress}
$beforeText=(& $old @g '--state-dir' $compare history --limit 100 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'Before history failed'};$before=$beforeText|ConvertFrom-Json
$afterText=(& $target @g '--state-dir' $restored history --limit 100 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'After history failed'};$after=$afterText|ConvertFrom-Json
$r=@{result='retained-checked';before_count=$before.Count;retained_count=$after.Count;retained_original_history_projection_preserved=((Projection $before)-ceq (Projection $after));private_snapshot_validated=$true;session_rows_exported=0;current_state_modified=$false}
$r|ConvertTo-Json|Set-Content (Join-Path $base 'daily-retained-comparison-summary.json') -Encoding utf8
$r|ConvertTo-Json
