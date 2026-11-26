$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$target=Join-Path $package 'asd-integrated-74a86959.exe';$old=Join-Path $package 'asd-focus-3654c0e.exe'
$compare=Join-Path $base 'daily-final-private-comparison-20261126'
if(Test-Path $compare){throw 'Fresh private comparison required'}
New-Item -ItemType Directory $compare|Out-Null
$g=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',(Join-Path $package 'data/runtime'),'--namespace','asd-alpha-8234c4e')
$backup=Join-Path $package 'data/rollback-integrated-20261126/metadata-v5.sqlite3'
$privateError=Join-Path $compare 'restore-error-private.txt'
& $target @g '--state-dir' $compare metadata restore $backup 1>$null 2>$privateError
$restoreExit=$LASTEXITCODE
if($restoreExit -ne 0){$errorText=Get-Content -Raw $privateError;@{result='clone-restore-failed';restore_exit=$restoreExit;destination_exists=(Test-Path (Join-Path $compare 'metadata.sqlite3'));destination_guard=$errorText.Contains('snapshot destination must not exist');locked=$errorText.Contains('locked');publication_error=$errorText.Contains('publish snapshot');session_content_exported=0}|ConvertTo-Json;return}
function Rows([string]$exe,[string]$state){$text=(& $exe @g '--state-dir' $state history --limit 100 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'Private bounded history read failed'};return @($text|ConvertFrom-Json)}
function Projection($rows){$selected=@($rows|ForEach-Object{$s=$_.session;[ordered]@{session=$s.session_id;run=$s.run_id;backend=$s.backend_id;command=$s.command;title=$s.title;title_revision=$s.title_revision;group=$s.group_id;group_revision=$s.group_revision;position=$s.order;position_revision=$s.order_revision}}|Sort-Object {$_.session},{$_.run});return ConvertTo-Json -InputObject $selected -Depth 6 -Compress}
$before=@(Rows $old $compare);$after=@(Rows $target (Join-Path $package 'data/state'))
$liveText=(& $target @g '--state-dir' (Join-Path $package 'data/state') list --json 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'Current live read failed'};$live=@($liveText|ConvertFrom-Json)
$unknown=@($live|Where-Object{$l=$_;!@($before|Where-Object{$_.session.session_id -ceq $l.session_id -and $_.session.run_id -ceq $l.run_id -and $_.session.backend_id -ceq $l.backend_id}).Count})
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$shellValid=0
foreach($label in @('A','B')){$path=Join-Path $package ('data/shell-'+$label+'.identity.json');if(Test-Path $path){$id=Get-Content -Raw $path|ConvertFrom-Json;$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;if($p -and $p.Path -ieq $hostInfo.shell.path -and $p.StartTime.ToUniversalTime().ToFileTimeUtc().ToString() -ceq ([string]$id.filetime)){$shellValid++}}}
$shell=New-Object -ComObject WScript.Shell;$links=@(Get-ChildItem ([Environment]::GetFolderPath('Desktop')) -Filter '*Session*Desk*.lnk' -File|Where-Object{$s=$shell.CreateShortcut($_.FullName);$s.Arguments.IndexOf((Join-Path $package 'start-integrated-74a86959.ps1'),[StringComparison]::OrdinalIgnoreCase)-ge 0})
$launchBody=[IO.File]::ReadAllText((Join-Path $package 'start-integrated-74a86959.ps1'))
$r=@{result='compared';bounded_history_before=$before.Count;bounded_history_after=$after.Count;history_projection_preserved=((Projection $before)-ceq (Projection $after));current_live_count=$live.Count;current_live_original_identity_match=($unknown.Count -eq 0);original_seeded_shell_identity_matches=$shellValid;updated_shortcuts=$links.Count;explicit_pet_and_system_notepad=($launchBody.Contains("'--pet' '--document-app' 'system-notepad'"));private_clone_restore_exit=$restoreExit;session_rows_exported=0;registry_writes=0;terminal_input_actions=0;database_content_exported=0}
$r|ConvertTo-Json|Set-Content (Join-Path $base 'daily-final-comparison-summary.json') -Encoding utf8
$r|ConvertTo-Json
