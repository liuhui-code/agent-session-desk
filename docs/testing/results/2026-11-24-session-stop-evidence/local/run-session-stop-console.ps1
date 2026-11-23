$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-stop-console-20261124'
if(Test-Path $root){throw 'Fresh root required'}
$product=Join-Path $base 'asd-session-stop.exe';$bundle=Join-Path $base 'terminal-bundle-20261123/bundle';$mux=Join-Path $bundle 'psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$expectedProduct='e0f72e83ec4bdac4115e47d49cf7494ccb71bbcc33382cacc5fb6a7bf0288204'
foreach($p in @(@($product,$expectedProduct),@((Join-Path $bundle 'OpenConsole.exe'),'b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775'),@($mux,'ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'),@($shell,'f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336'),@((Join-Path $bundle 'conpty.dll'),'e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492'))){if((Get-FileHash $p[0]).Hash.ToLowerInvariant() -cne $p[1]){throw 'Pinned binary mismatch'}}
New-Item -ItemType Directory $root|Out-Null;$ns='asd-session-stop';$driver='asd-session-stop-driver';$data=Join-Path $root 'driver';New-Item -ItemType Directory $data|Out-Null
$g=@('--terminal-bundle',$bundle,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $root 'state'))
foreach($entry in @(Get-ChildItem Env:|Where-Object{$_.Name -like 'PSMUX_*' -or $_.Name -in @('TMUX','TMUX_PANE','TMUX_TMPDIR')})){Remove-Item ('Env:'+$entry.Name)}
$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1';$env:PSMUX_DATA_DIR=$data;$env:PSMUX_CONPTY_DIR=$bundle;$env:PSMUX_CONFIG_FILE=Join-Path $root 'psmux.conf';[IO.File]::WriteAllText($env:PSMUX_CONFIG_FILE,"set -g warm off`nset -g status off`nset -g history-limit 2000`n")
$events=[Collections.Generic.List[object]]::new();$script:stopCaptureNumber=0;$sessions=@();$shells=@();$script:stopManagers=@();$script:stopServers=@();$script:stopIdentities=@();$script:stopTarget=$null;$success=$false;$forced=0
function Event($Kind,$Details){$events.Add(@{kind=$Kind;utc=[DateTime]::UtcNow.ToString('o');details=$Details});$events|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'events.json') -Encoding utf8}
function Mux([string[]]$Arguments){$r=(& $mux -L $driver @Arguments|Out-String);if($LASTEXITCODE -ne 0){throw ('Driver failed: '+($Arguments -join ' '))};return $r.TrimEnd("`r","`n")}
function Text([string]$Value){$null=Mux @('send-keys','-t',$script:stopTarget,'-l',$Value);Event 'text' @{value=$Value}}
function Key([string]$Value){$null=Mux @('send-keys','-t',$script:stopTarget,$Value);Event 'key' @{value=$Value}}
function Screen{return Mux @('capture-pane','-t',$script:stopTarget,'-p')}
function Await([string]$Value){$until=[DateTime]::UtcNow.AddSeconds(10);do{$s=Screen;if($s.Contains($Value)){return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);[IO.File]::WriteAllText((Join-Path $root 'failed-screen.txt'),$s);throw ('Expected screen missing: '+$Value+'; no action replayed')}
function AwaitManagementStable {$until=[DateTime]::UtcNow.AddSeconds(10);do{$s=Screen;if($s.Contains('管理模式') -and $s -notmatch '正在保存|未确认|移动到分组|创建分组|创建子组|重命名分组'){return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);throw 'Stable management frame missing; no action replayed'}
function Capture([string]$Name){$script:stopCaptureNumber++;$v=@{number=$script:stopCaptureNumber;name=$Name;text=(Screen)};$v|ConvertTo-Json|Set-Content (Join-Path $root ('capture-'+($script:stopCaptureNumber).ToString('000')+'.json')) -Encoding utf8}
function Rows{$s=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'JSON list failed'};return ($s|ConvertFrom-Json)}
function WaitOrder([string[]]$Expected){$until=[DateTime]::UtcNow.AddSeconds(10);do{$rows=@(Rows);if(($rows.session_id -join ',') -ceq ($Expected -join ',')){return $rows};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);throw 'Persistent stop not observed; key not replayed'}
function SnapshotRows([string]$Name,[object[]]$Rows){$Rows|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root ($Name+'.json')) -Encoding utf8}
function Register([Diagnostics.Process]$Process,[string]$Kind){$handle=$Process.Handle;$script:stopIdentities+=,@{pid=$Process.Id;filetime=$Process.StartTime.ToUniversalTime().ToFileTimeUtc();path=$Process.Path;kind=$Kind;handle_retained=($handle -ne [IntPtr]::Zero)};$script:stopIdentities|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'identities.json') -Encoding utf8}
function ReadSharedText([string]$Path){$stream=[IO.FileStream]::new($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete));try{$reader=[IO.StreamReader]::new($stream);try{return $reader.ReadToEnd()}finally{$reader.Dispose()}}finally{$stream.Dispose()}}
function ReadFixtureIdentity([string]$Path){$until=[DateTime]::UtcNow.AddSeconds(10);do{try{if(Test-Path $Path){$v=(ReadSharedText $Path)|ConvertFrom-Json;if($v.pid -and $v.filetime -and $v.path){return $v}}}catch{};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);throw 'Fixture identity not fully observed; start was not retried'}
function StartUI([string]$Name){$driverData=Join-Path $data $Name;if(Test-Path $driverData){throw 'Fresh UI driver data required'};New-Item -ItemType Directory $driverData|Out-Null;$env:PSMUX_DATA_DIR=$driverData;$null=Mux (@('new-session','-d','-s',$Name,'-c',$root,'-x','120','-y','32','--',$product)+$g);$pane=(Mux @('list-panes','-t',$Name,'-F','#{pane_id}')).Trim();if($pane -notmatch '^%\d+$'){throw 'Exact driver pane missing'};$script:stopTarget=$Name+':.'+$pane;$managerPid=[int](Mux @('display-message','-p','-t',$script:stopTarget,'#{pane_pid}'));$p=Get-Process -Id $managerPid;if($p.Path -ne $product){throw 'Manager image mismatch'};Register $p ('manager-'+$Name);$script:stopManagers+=,$p
$files=@(Get-ChildItem $driverData -Filter '*.pid');if($files.Count -ne 1){throw 'Driver registry mismatch'};$parts=(ReadSharedText $files[0].FullName).Trim().Split(':');$server=Get-Process -Id ([int]$parts[0]);if($server.Path -ne $mux -or $server.StartTime.ToUniversalTime().ToFileTimeUtc() -ne [long]$parts[1]){throw 'Original driver mismatch'};Register $server ('driver-'+$Name);$script:stopServers+=,$server;$mods=@($server.Modules|Where-Object{$_.FileName -eq (Join-Path $bundle 'conpty.dll')});if($mods.Count -ne 1){throw 'Expected driver DLL not observed'};@{ui=$Name;pid=$server.Id;module=$mods[0].FileName;sha256=(Get-FileHash $mods[0].FileName).Hash}|ConvertTo-Json|Set-Content (Join-Path $root ($Name+'-host.json')) -Encoding utf8}
function StopUI {Await '管理模式';Key 'q';$p=$script:stopManagers[-1];if(!$p.WaitForExit(15000)-or $null -eq $p.ExitCode -or $p.ExitCode -ne 0){throw 'Original manager exit not confirmed zero'};Event 'manager-exited' @{pid=$p.Id;exit_code=$p.ExitCode};$null=$script:stopServers[-1].WaitForExit(10000);$script:stopTarget=$null}
function AwaitOutput([string]$Marker){$until=[DateTime]::UtcNow.AddSeconds(10);$pattern='(?m)│\s*'+[regex]::Escape($Marker)+'\s*│\s*$';do{$s=Screen;if($s -match $pattern){return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);[IO.File]::WriteAllText((Join-Path $root 'failed-screen.txt'),$s);throw 'Explicit original-shell output missing; Enter not replayed'}
try{
$fixture=Join-Path $root 'shell.ps1'
@'
param([string]$label)
function global:prompt { '> ' }
Set-PSReadLineOption -HistorySavePath (Join-Path $PSScriptRoot ($label+'.history.txt'))
$p=Get-Process -Id $PID
$identity=Join-Path $PSScriptRoot ($label+'.identity.json')
$json=@{label=$label;pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json
[IO.File]::WriteAllText(($identity+'.pending'),$json,[Text.UTF8Encoding]::new($false));[IO.File]::Move(($identity+'.pending'),$identity)
Clear-Host
'STOP_READY '+$label
'@|Set-Content $fixture -Encoding utf8
foreach($name in @('A','B')){$v=(& $product @g start --name ('Stop '+$name) --cwd $root -- $shell -NoLogo -NoProfile -NoExit -File $fixture $name|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0){throw 'Shell start failed'};$sessions+=,$v;$sessions|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'sessions-before.json') -Encoding utf8}
foreach($name in @('A','B')){$id=ReadFixtureIdentity (Join-Path $root ($name+'.identity.json'));$p=Get-Process -Id $id.pid;if($p.Path -ne $shell -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime){throw 'Original shell differs'};Register $p ('shell-'+$name);$shells+=,$p}
& $product @g history|Set-Content (Join-Path $root 'history-before.json') -Encoding utf8
StartUI 'ui1';Await 'STOP_READY A';Text "Write-Output ('STOP_A:'+`$PID)";Await "Write-Output ('STOP_A:'+`$PID)";Capture 'A-original-draft'
Key 'C-PageDown';Await 'STOP_READY B';Text "Write-Output ('STOP_B:'+`$PID)";Await "Write-Output ('STOP_B:'+`$PID)";Capture 'B-original-draft'
Key 'C-PageUp';Await "Write-Output ('STOP_A:'+`$PID)";Key 'F6';Await '管理模式';Key 'Down';Text 'd';Await '停止会话';Await $sessions[1].run_id;Capture 'B-selected-A-active-default-cancel'
Key 'Enter';Await '管理模式';Await "Write-Output ('STOP_A:'+`$PID)";if($shells[0].HasExited -or $shells[1].HasExited){throw 'Default cancel stopped shell'};Capture 'cancel-preserved-both'
Text 'd';Await '停止会话';$null=Mux @('resize-window','-t',$script:stopTarget,'-x','120','-y','10');Await '请放大窗口';Text 'y';Await '请放大窗口';if(@(Rows).Count -ne 2 -or $shells[1].HasExited){throw 'Unreadable review stopped target'};Capture 'narrow-review-refused'
$null=Mux @('resize-window','-t',$script:stopTarget,'-x','120','-y','32');Await $sessions[1].run_id;Capture 'B-full-review-restored';Text 'y'
$until=[DateTime]::UtcNow.AddSeconds(10);do{$rows=@(Rows);if($rows.Count -eq 1 -and $rows[0].session_id -eq $sessions[0].session_id){break};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until)
if($rows.Count -ne 1 -or $rows[0].run_id -ne $sessions[0].run_id -or !$shells[1].WaitForExit(10000) -or $shells[0].HasExited){throw 'Selected B stop did not preserve original A'}
SnapshotRows 'after-B-stop' $rows;Await '管理模式';Await "Write-Output ('STOP_A:'+`$PID)";Capture 'B-stopped-A-draft-preserved'
$history=(& $product @g history|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0){throw 'History query failed'};SnapshotRows 'history-after-B-stop' @($history)
$hB=@($history|Where-Object{$_.session.session_id -eq $sessions[1].session_id});$hA=@($history|Where-Object{$_.session.session_id -eq $sessions[0].session_id});if($hB.Count -ne 1 -or !$hB[0].stop_confirmed_at -or $hA.Count -ne 1 -or $hA[0].stop_confirmed_at){throw 'Stopped history confused with live A'}
Key 'C-PageDown';Await '目标终端当前不可交互';Capture 'B-view-retained';Text 'x';Await "Write-Output ('STOP_A:'+`$PID)";Capture 'B-view-closed-A-remains'
Text 'd';Await '停止会话';Await $sessions[0].run_id;Capture 'A-reviewed';Key 'Left';Key 'Enter'
$until=[DateTime]::UtcNow.AddSeconds(10);do{$rows=@(Rows);if($rows.Count -eq 0){break};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until)
if(!$shells[0].WaitForExit(10000) -or $rows.Count -ne 0){throw 'Active A stop missing'};Capture 'A-stopped-view-retained'
$history=(& $product @g history|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or @($history).Count -ne 2 -or @($history|Where-Object{!$_.stop_confirmed_at}).Count -ne 0){throw 'Original stopped history missing'};SnapshotRows 'history-final' @($history)
StopUI
StartUI 'ui2';Await '暂无打开的 Tab';Await '管理模式';Capture 'reopened-no-live-sessions';Text 'h';Await '历史记录';Capture 'reopened-history-preserved';Key 'Esc';Await '管理模式';StopUI
$success=$true
}catch{Event 'failure' @{error=$_.Exception.Message};throw}finally{
$remaining=@(Rows);foreach($v in $sessions){$found=@($remaining|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id -and $_.backend_id -eq $v.backend_id});if($found.Count -eq 1){& $product @g stop $v.session_id|Set-Content (Join-Path $root ('cleanup-stop-'+$v.session_id+'.json')) -Encoding utf8}}
if($script:stopTarget -and $script:stopManagers.Count -gt 0 -and !$script:stopManagers[-1].HasExited){$forced++;$name=$script:stopTarget.Split(':')[0];try{$null=Mux @('kill-session','-t',$name)}catch{Event 'cleanup-error' @{error=$_.Exception.Message}}}
foreach($p in @($shells)+@($script:stopManagers)+@($script:stopServers)){$null=$p.WaitForExit(10000)}
$ended=@();foreach($id in $script:stopIdentities){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$alive=$false;if($p){$alive=$p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $id.filetime};$ended+=,@{pid=$id.pid;filetime=$id.filetime;kind=$id.kind;original_ended=(!$alive)}}
$r=@{utc=[DateTime]::UtcNow.ToString('o');result=$(if($success){'passed'}else{'failed'});captures=$script:stopCaptureNumber;events=$events.Count;product_sha256=(Get-FileHash $product).Hash;windows=[Environment]::OSVersion.VersionString;originals=$ended;forced_product_ends=$forced;clipboard_api_calls=0;desktop_surface='none: isolated ConPTY';remaining_exact_product=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $product}).Count}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8;$r|ConvertTo-Json -Depth 8
}
