$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tabs-ui-20261123'
$product=Join-Path $base 'tabs-ui-final.exe';$backend=Join-Path $base 'backend/psmux.exe'
$g=@('--backend',$backend,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-tabs-ui','--state-dir',(Join-Path $root 'state'))
$rowsText=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'List failed'};$rows=$rowsText|ConvertFrom-Json
$original=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json
$stops=@()
foreach($s in $rows){
 $expected=@($original|Where-Object{$_.session_id -eq $s.session_id -and $_.run_id -eq $s.run_id})
 if($expected.Count -ne 1 -and !($s.title -eq '普通终端 C' -and $s.command -eq (Join-Path $base 'pwsh7/pwsh.exe'))){throw 'Unexpected target; cleanup refused'}
 $output=(& $product @g stop $s.session_id|Out-String);$code=$LASTEXITCODE
 $stops+=,@{session_id=$s.session_id;run_id=$s.run_id;code=$code;output=$output}
 if($code -ne 0){throw 'Owned session stop failed'}
}
$after=(& $product @g list --json|Out-String).Trim();if($LASTEXITCODE -ne 0 -or $after -ne '[]'){throw 'Owned sessions remain'}
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
$active=@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName})
if($active.Count -ne 0){throw 'Private terminal remains; settings restore refused'}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';$backup=Join-Path $root 'terminal-settings.before'
if((Get-FileHash $settings).Hash -ne '1F0E4A595CB909336D4E0D942AD1ABA37B919F9B25208F659BE0AABEDEAA12C2'){throw 'Private settings changed externally; restore refused'}
[IO.File]::Copy($backup,$settings,$true)
if((Get-FileHash $settings).Hash -ne (Get-FileHash $backup).Hash){throw 'Settings restore mismatch'}
$tasks=@()
foreach($name in @('AsdTabsDesktop20261123','AsdTabsDesktop2_20261123','AsdTabsPasteProbe20261123','AsdTabsScanProbe20261123')){
 $task=Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue
 if($task){if($task.State -eq 'Running'){throw "Owned task still running: $name"};Unregister-ScheduledTask -TaskName $name -Confirm:$false;$tasks+=,$name}
}
$result=@{utc=[DateTime]::UtcNow.ToString('o');before=$rows;stops=$stops;after_json=$after;private_terminal_count=$active.Count;settings_restored_sha256=(Get-FileHash $settings).Hash;removed_tasks=$tasks;result='passed'}
$result|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'cleanup.json') -Encoding utf8
$result|ConvertTo-Json -Depth 8
