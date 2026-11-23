$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-parity-desktop-20261123';$product=Join-Path $base 'asd-main-parity.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe'
$task='AsdMainParityDesktop20261123';if((Get-ScheduledTask $task).State -eq 'Running'){throw 'Driver still running'}
$ids=([IO.File]::ReadAllText((Join-Path $root 'blocked-cleanup-identities.json'))|ConvertFrom-Json).identities
function Exact($id){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;if(!$p -or $p.Path -ne $id.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.SessionId -ne $id.session){return $null};return $p}
foreach($id in $ids){if(!(Exact $id)){throw 'Registered original changed before cleanup'}}
$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-main-parity','--state-dir',(Join-Path $root 'state'))
$owned=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json;$rows=(& $product @g list --json|Out-String)|ConvertFrom-Json
if($LASTEXITCODE -ne 0 -or @($rows).Count -ne 2){throw 'Unexpected live set'}
foreach($v in $rows){if(@($owned|Where-Object{$_.session_id -ceq $v.session_id -and $_.run_id -ceq $v.run_id -and $_.backend_id -ceq $v.backend_id}).Count -ne 1){throw 'Unowned live target'}}
$env:PSMUX_DATA_DIR=Join-Path (Join-Path $root 'runtime') 'asd-main-parity';$env:PSMUX_CONFIG_FILE=Join-Path $env:PSMUX_DATA_DIR 'psmux-m0.conf';$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1'
$captures=@();foreach($v in $owned){$text=(& $backend -L 'asd-main-parity' capture-pane -t $v.backend_id -p -S -200|Out-String);if($LASTEXITCODE -ne 0){throw 'Read-only setup capture failed'};$numbers=@(($text -split '\r?\n')|ForEach-Object{if($_ -match '^history line (\d+)$'){[int]$Matches[1]}});if(($numbers -join ',') -cne ((1..40) -join ',') -or $text -match 'MAIN_[AB]:'){throw 'Setup output differs or unexpected flow input'};$captures+=,@{session_id=$v.session_id;run_id=$v.run_id;title=$v.title;numbered_ids=$numbers;text=$text;scope='Setup shell only; no desktop actions were accepted'}}
$captures|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'setup-captures.json') -Encoding utf8
$stops=@();foreach($v in $owned){$text=(& $product @g stop $v.session_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Owned stop failed'};$stops+=,@{session_id=$v.session_id;run_id=$v.run_id;result=$text}}
$stops|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'blocked-owned-stops.json') -Encoding utf8
$name='AsdMainParityClose20261123';if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Cleanup task already exists'}
$code=@'
$ErrorActionPreference='Stop';$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-parity-desktop-20261123'
$ids=([IO.File]::ReadAllText((Join-Path $root 'blocked-cleanup-identities.json'))|ConvertFrom-Json).identities
function Exact($id){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;if(!$p -or $p.Path -ne $id.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime -or $p.SessionId -ne $id.session){return $null};return $p}
$terminal=@($ids|Where-Object{$_.kind -eq 'terminal'})[0];$p=Exact $terminal;$close=$false;if($p){$close=$p.CloseMainWindow()}
$until=[DateTime]::UtcNow.AddSeconds(8);do{$live=@($ids|Where-Object{Exact $_});if(!$live.Count){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
$forced=@();foreach($id in $ids){$p=Exact $id;if($p){$forced+=,$id;Stop-Process -Id $p.Id -ErrorAction Stop}}
$until=[DateTime]::UtcNow.AddSeconds(5);do{$live=@($ids|Where-Object{Exact $_});if(!$live.Count){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
@{utc=[DateTime]::UtcNow.ToString('o');close_window_requested=$close;forced=$forced;remaining_registered=$live.Count;scope='Exact registered originals only; cleanup after blocked desktop is not normal UI q or desktop acceptance'}|ConvertTo-Json -Depth 7|Set-Content (Join-Path $root 'blocked-ui-close.json') -Encoding utf8
'@
$a=New-ScheduledTaskAction -Execute powershell.exe -Argument ('-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code)));$p=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask $name -Action $a -Principal $p -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 1))|Out-Null;Start-ScheduledTask $name
$file=Join-Path $root 'blocked-ui-close.json';$until=[DateTime]::UtcNow.AddSeconds(25);while(!(Test-Path $file) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $file)){throw 'Exact cleanup result not observed'}
Get-Content $file -Raw -Encoding utf8
