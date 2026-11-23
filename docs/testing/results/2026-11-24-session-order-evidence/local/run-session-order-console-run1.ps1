$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-order-console-20261124'
if(Test-Path $root){throw 'Fresh root required'}
$product=Join-Path $base 'asd-session-order.exe';$legacy=Join-Path $base 'asd-long-field-final.exe';$bundle=Join-Path $base 'terminal-bundle-20261123/bundle';$mux=Join-Path $bundle 'psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$expectedProduct='062dca624139610b2134ae3885db4c8ab8d0f8f92e179a3b7d211ec7dd2b87ec'
foreach($p in @(@($product,$expectedProduct),@($legacy,'2f022aaddec1aa8fc247c06e9557a97a7e2af9a8d227e7dd3e47dca6719e10a8'),@($mux,'ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'),@($shell,'f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336'),@((Join-Path $bundle 'conpty.dll'),'e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492'))){if((Get-FileHash $p[0]).Hash.ToLowerInvariant() -cne $p[1]){throw 'Pinned binary mismatch'}}
New-Item -ItemType Directory $root|Out-Null;$ns='asd-session-order';$driver='asd-session-order-driver';$data=Join-Path $root 'driver';New-Item -ItemType Directory $data|Out-Null
$g=@('--terminal-bundle',$bundle,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $root 'state'))
foreach($entry in @(Get-ChildItem Env:|Where-Object{$_.Name -like 'PSMUX_*' -or $_.Name -in @('TMUX','TMUX_PANE','TMUX_TMPDIR')})){Remove-Item ('Env:'+$entry.Name)}
$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1';$env:PSMUX_DATA_DIR=$data;$env:PSMUX_CONPTY_DIR=$bundle;$env:PSMUX_CONFIG_FILE=Join-Path $root 'psmux.conf';[IO.File]::WriteAllText($env:PSMUX_CONFIG_FILE,"set -g warm off`nset -g status off`nset -g history-limit 2000`n")
$events=[Collections.Generic.List[object]]::new();$script:orderCaptureNumber=0;$sessions=@();$shells=@();$script:orderManagers=@();$script:orderServers=@();$script:orderIdentities=@();$script:orderTarget=$null;$success=$false;$forced=0
function Event($Kind,$Details){$events.Add(@{kind=$Kind;utc=[DateTime]::UtcNow.ToString('o');details=$Details});$events|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'events.json') -Encoding utf8}
function Mux([string[]]$Arguments){$r=(& $mux -L $driver @Arguments|Out-String);if($LASTEXITCODE -ne 0){throw ('Driver failed: '+($Arguments -join ' '))};return $r.TrimEnd("`r","`n")}
function Text([string]$Value){$null=Mux @('send-keys','-t',$script:orderTarget,'-l',$Value);Event 'text' @{value=$Value}}
function Key([string]$Value){$null=Mux @('send-keys','-t',$script:orderTarget,$Value);Event 'key' @{value=$Value}}
function Screen{return Mux @('capture-pane','-t',$script:orderTarget,'-p')}
function Await([string]$Value){$until=[DateTime]::UtcNow.AddSeconds(10);do{$s=Screen;if($s.Contains($Value)){return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);[IO.File]::WriteAllText((Join-Path $root 'failed-screen.txt'),$s);throw ('Expected screen missing: '+$Value+'; no action replayed')}
function Capture([string]$Name){$script:orderCaptureNumber++;$v=@{number=$script:orderCaptureNumber;name=$Name;text=(Screen)};$v|ConvertTo-Json|Set-Content (Join-Path $root ('capture-'+($script:orderCaptureNumber).ToString('000')+'.json')) -Encoding utf8}
function Rows{$s=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'JSON list failed'};return ($s|ConvertFrom-Json)}
function WaitOrder([string[]]$Expected){$until=[DateTime]::UtcNow.AddSeconds(10);do{$rows=@(Rows);if(($rows.session_id -join ',') -ceq ($Expected -join ',')){return $rows};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);throw 'Persistent order not observed; key not replayed'}
function SnapshotRows([string]$Name,[object[]]$Rows){$Rows|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root ($Name+'.json')) -Encoding utf8}
function Register([Diagnostics.Process]$Process,[string]$Kind){$handle=$Process.Handle;$script:orderIdentities+=,@{pid=$Process.Id;filetime=$Process.StartTime.ToUniversalTime().ToFileTimeUtc();path=$Process.Path;kind=$Kind;handle_retained=($handle -ne [IntPtr]::Zero)};$script:orderIdentities|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'identities.json') -Encoding utf8}
function StartUI([string]$Name){$null=Mux (@('new-session','-d','-s',$Name,'-c',$root,'-x','120','-y','32','--',$product)+$g);$pane=(Mux @('list-panes','-t',$Name,'-F','#{pane_id}')).Trim();if($pane -notmatch '^%\d+$'){throw 'Exact driver pane missing'};$script:orderTarget=$Name+':.'+$pane;$managerPid=[int](Mux @('display-message','-p','-t',$script:orderTarget,'#{pane_pid}'));$p=Get-Process -Id $managerPid;if($p.Path -ne $product){throw 'Manager image mismatch'};Register $p ('manager-'+$Name);$script:orderManagers+=,$p
$files=@(Get-ChildItem $data -Filter '*.pid');if($files.Count -ne 1){throw 'Driver registry mismatch'};$parts=([IO.File]::ReadAllText($files[0].FullName)).Trim().Split(':');$server=Get-Process -Id ([int]$parts[0]);if($server.Path -ne $mux -or $server.StartTime.ToUniversalTime().ToFileTimeUtc() -ne [long]$parts[1]){throw 'Original driver mismatch'};Register $server ('driver-'+$Name);$script:orderServers+=,$server;$mods=@($server.Modules|Where-Object{$_.FileName -eq (Join-Path $bundle 'conpty.dll')});if($mods.Count -ne 1){throw 'Expected driver DLL not observed'};@{ui=$Name;pid=$server.Id;module=$mods[0].FileName;sha256=(Get-FileHash $mods[0].FileName).Hash}|ConvertTo-Json|Set-Content (Join-Path $root ($Name+'-host.json')) -Encoding utf8}
function StopUI {Await '管理模式';Key 'q';$p=$script:orderManagers[-1];if(!$p.WaitForExit(15000)-or $null -eq $p.ExitCode -or $p.ExitCode -ne 0){throw 'Original manager exit not confirmed zero'};Event 'manager-exited' @{pid=$p.Id;exit_code=$p.ExitCode};$null=$script:orderServers[-1].WaitForExit(10000);$script:orderTarget=$null}
function AwaitOutput([string]$Marker){$until=[DateTime]::UtcNow.AddSeconds(10);$pattern='(?m)│\s*'+[regex]::Escape($Marker)+'\s*│\s*$';do{$s=Screen;if($s -match $pattern){return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);[IO.File]::WriteAllText((Join-Path $root 'failed-screen.txt'),$s);throw 'Explicit original-shell output missing; Enter not replayed'}
try{
$fixture=Join-Path $root 'shell.ps1'
@'
param([string]$label)
function global:prompt { '> ' }
Set-PSReadLineOption -HistorySavePath (Join-Path $PSScriptRoot ($label+'.history.txt'))
$p=Get-Process -Id $PID
@{label=$label;pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content (Join-Path $PSScriptRoot ($label+'.identity.json')) -Encoding utf8
Clear-Host
'ORDER_READY '+$label
'@|Set-Content $fixture -Encoding utf8
foreach($name in @('A','B')){$v=(& $legacy @g start --name ('Order '+$name) --cwd $root -- $shell -NoLogo -NoProfile -NoExit -File $fixture $name|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0){throw 'Legacy shell start failed'};$sessions+=,$v;$sessions|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'sessions-before.json') -Encoding utf8}
$legacyCheck=(& $legacy @g metadata check|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $legacyCheck.schema_version -ne 4){throw 'Actual v4 source missing'};$legacyCheck|ConvertTo-Json|Set-Content (Join-Path $root 'metadata-before.json') -Encoding utf8
& $legacy @g history|Set-Content (Join-Path $root 'history-before.json') -Encoding utf8
$until=[DateTime]::UtcNow.AddSeconds(10)
foreach($name in @('A','B')){$file=Join-Path $root ($name+'.identity.json');while(!(Test-Path $file)-and[DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 80};$id=[IO.File]::ReadAllText($file)|ConvertFrom-Json;$p=Get-Process -Id $id.pid;if($p.Path -ne $shell -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime){throw 'Original shell differs'};Register $p ('shell-'+$name);$shells+=,$p}
StartUI 'ui1';Await 'ORDER_READY A';Text "Write-Output ('ORDER_A:'+`$PID)";Await "Write-Output ('ORDER_A:'+`$PID)";Capture 'A-draft'
Key 'C-PageDown';Await 'ORDER_READY B';Text "Write-Output ('ORDER_B:'+`$PID)";Await "Write-Output ('ORDER_B:'+`$PID)";Capture 'B-draft';Key 'F6';Await '管理模式'
Text 'K';$rows=@(WaitOrder @($sessions[1].session_id,$sessions[0].session_id));SnapshotRows 'order-up' $rows;Capture 'selected-B-up'
Text 'J';$rows=@(WaitOrder @($sessions[0].session_id,$sessions[1].session_id));SnapshotRows 'order-down' $rows;Capture 'selected-B-down'
Text 'K';$rows=@(WaitOrder @($sessions[1].session_id,$sessions[0].session_id));SnapshotRows 'order-final' $rows;Capture 'selected-B-up-again';StopUI
foreach($p in $shells){if($p.HasExited){throw 'UI exit stopped original shell'}}
$db=Join-Path $root 'state/metadata.sqlite3';$hashBefore=(Get-FileHash $db).Hash;$previousEAP=$ErrorActionPreference;$ErrorActionPreference='Continue'
& $legacy @g metadata check > (Join-Path $root 'legacy-v5-check.out') 2> (Join-Path $root 'legacy-v5-check.err');$legacyCode=$LASTEXITCODE;$ErrorActionPreference=$previousEAP
$hashAfter=(Get-FileHash $db).Hash;if($legacyCode -eq 0 -or $hashBefore -cne $hashAfter){throw 'Legacy binary accepted/changed v5'}
@{exit_code=$legacyCode;sha256_before=$hashBefore;sha256_after=$hashAfter}|ConvertTo-Json|Set-Content (Join-Path $root 'legacy-v5-rejection.json') -Encoding utf8
$check=(& $product @g metadata check|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $check.schema_version -ne 5){throw 'Final v5 metadata check failed'};$check|ConvertTo-Json|Set-Content (Join-Path $root 'metadata-after.json') -Encoding utf8
& $product @g history|Set-Content (Join-Path $root 'history-after.json') -Encoding utf8
$backup=Join-Path $root 'backup.sqlite3';& $product @g metadata backup $backup|Set-Content (Join-Path $root 'backup-result.json') -Encoding utf8;if($LASTEXITCODE -ne 0){throw 'Backup failed'}
$restored=Join-Path $root 'restored';$rg=@('--terminal-bundle',$bundle,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace',$ns,'--state-dir',$restored)
& $product @rg metadata restore $backup|Set-Content (Join-Path $root 'restore-result.json') -Encoding utf8;if($LASTEXITCODE -ne 0){throw 'Restore failed'}
$restoredRows=(& $product @rg list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or ($restoredRows.session_id -join ',') -cne ($rows.session_id -join ',')){throw 'Restored order differs'};SnapshotRows 'restored-order' $restoredRows
StartUI 'ui2';Await 'ORDER_READY B';Await "Write-Output ('ORDER_B:'+`$PID)";Capture 'reopened-B-first'
Key 'F6';Await '管理模式';Key 'Down';Key 'Enter';Await '终端模式';Await "Write-Output ('ORDER_A:'+`$PID)";Capture 'reopened-A-original-draft';Key 'Enter';AwaitOutput ('ORDER_A:'+$shells[0].Id);Capture 'A-explicit-output'
Key 'F6';Await '管理模式';Key 'C-PageUp';Key 'F6';Await '终端模式';Await "Write-Output ('ORDER_B:'+`$PID)";Key 'Enter';AwaitOutput ('ORDER_B:'+$shells[1].Id);Capture 'B-explicit-output';Key 'F6';Await '管理模式';StopUI
$after=@(Rows);SnapshotRows 'sessions-after' $after
foreach($v in $sessions){$found=@($after|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id -and $_.backend_id -eq $v.backend_id});if($found.Count -ne 1){throw 'Original Session/Run/backend changed'}}
$success=$true
}catch{Event 'failure' @{error=$_.Exception.Message};throw}finally{
foreach($v in $sessions){& $product @g stop $v.session_id|Set-Content (Join-Path $root ('stop-'+$v.session_id+'.json')) -Encoding utf8}
if($script:orderTarget -and $script:orderManagers.Count -gt 0 -and !$script:orderManagers[-1].HasExited){$forced++;$name=$script:orderTarget.Split(':')[0];try{$null=Mux @('kill-session','-t',$name)}catch{Event 'cleanup-error' @{error=$_.Exception.Message}}}
foreach($p in @($shells)+@($script:orderManagers)+@($script:orderServers)){$null=$p.WaitForExit(10000)}
$ended=@();foreach($id in $script:orderIdentities){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$alive=$false;if($p){$alive=$p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $id.filetime};$ended+=,@{pid=$id.pid;filetime=$id.filetime;kind=$id.kind;original_ended=(!$alive)}}
$r=@{utc=[DateTime]::UtcNow.ToString('o');result=$(if($success){'passed'}else{'failed'});captures=$script:orderCaptureNumber;events=$events.Count;product_sha256=(Get-FileHash $product).Hash;windows=[Environment]::OSVersion.VersionString;originals=$ended;forced_product_ends=$forced;clipboard_api_calls=0;desktop_surface='none: isolated ConPTY';remaining_exact_product=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $product}).Count}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8;$r|ConvertTo-Json -Depth 8
}
