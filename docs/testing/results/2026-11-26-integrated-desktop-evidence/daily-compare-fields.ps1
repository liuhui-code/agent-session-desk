$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$target=Join-Path $package 'asd-integrated-74a86959.exe';$old=Join-Path $package 'asd-focus-3654c0e.exe'
$compare=Join-Path $base 'daily-final-private-comparison-20261126'
$g=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',(Join-Path $package 'data/runtime'),'--namespace','asd-alpha-8234c4e')
if(!(Test-Path (Join-Path $compare 'metadata.sqlite3'))){throw 'Validated comparison clone required'}
function Projection($rows){$selected=@($rows|ForEach-Object{$s=$_.session;[ordered]@{session=$s.session_id;run=$s.run_id;backend=$s.backend_id;command=$s.command;title=$s.title;title_revision=$s.title_revision;group=$s.group_id;group_revision=$s.group_revision;position=$s.order;position_revision=$s.order_revision}}|Sort-Object {$_.session},{$_.run});return ConvertTo-Json -InputObject $selected -Depth 6 -Compress}
$beforeText=(& $old @g '--state-dir' $compare history --limit 100 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'Before history failed'};$before=$beforeText|ConvertFrom-Json
$afterText=(& $target @g '--state-dir' (Join-Path $package 'data/state') history --limit 100 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'After history failed'};$after=$afterText|ConvertFrom-Json
$liveText=(& $target @g '--state-dir' (Join-Path $package 'data/state') list --json 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'Current live read failed'};$live=$liveText|ConvertFrom-Json
$unknown=@($live|Where-Object{$l=$_;!@($before|Where-Object{$_.session.session_id -ceq $l.session_id -and $_.session.run_id -ceq $l.run_id -and $_.session.backend_id -ceq $l.backend_id}).Count})
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$shellValid=0
foreach($label in @('A','B')){$path=Join-Path $package ('data/shell-'+$label+'.identity.json');if(Test-Path $path){$id=Get-Content -Raw $path|ConvertFrom-Json;$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;if($p -and $p.Path -ieq $hostInfo.shell.path -and $p.StartTime.ToUniversalTime().ToFileTimeUtc().ToString() -ceq ([string]$id.filetime)){$shellValid++}}}
$shell=New-Object -ComObject WScript.Shell;$links=@(Get-ChildItem ([Environment]::GetFolderPath('Desktop')) -Filter '*Session*Desk*.lnk' -File|Where-Object{$s=$shell.CreateShortcut($_.FullName);$encoded=($s.Arguments -split ' ')[-1];try{$body=[Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($encoded));return $body.IndexOf((Join-Path $package 'start-integrated-74a86959.ps1'),[StringComparison]::OrdinalIgnoreCase)-ge 0}catch{return $false}})
$launchBody=[IO.File]::ReadAllText((Join-Path $package 'start-integrated-74a86959.ps1'))
$fields=@('session_id','run_id','backend_id','command','title','title_revision','group_id','group_revision','order','order_revision')
$diff=@{};$missing=0;$duplicate=0
foreach($b in $before){$matches=@($after|Where-Object{$_.session.session_id -ceq $b.session.session_id -and $_.session.run_id -ceq $b.session.run_id});if($matches.Count -ne 1){$missing++;continue};foreach($field in $fields){$v1=ConvertTo-Json -InputObject $b.session.$field -Compress;$v2=ConvertTo-Json -InputObject $matches[0].session.$field -Compress;if($v1 -cne $v2){if(!$diff.ContainsKey($field)){$diff[$field]=0};$diff[$field]++}}}
@{result='field-diff';before_count=$before.Count;after_count=$after.Count;missing_or_nonunique_original_records=$missing;changed_field_counts=$diff;session_values_exported=0;terminal_input_actions=0}|ConvertTo-Json -Depth 4
