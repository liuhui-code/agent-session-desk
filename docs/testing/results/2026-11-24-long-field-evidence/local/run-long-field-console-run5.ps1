$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'long-field-console-run5-20261124'
if(Test-Path $root){throw 'Fresh console root required'}
$product=Join-Path $base 'asd-long-field-final.exe';$bundle=Join-Path $base 'terminal-bundle-20261123/bundle';$mux=Join-Path $bundle 'psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
foreach($pair in @(@($product,'2f022aaddec1aa8fc247c06e9557a97a7e2af9a8d227e7dd3e47dca6719e10a8'),@($mux,'ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'),@($shell,'f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336'),@((Join-Path $bundle 'conpty.dll'),'e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492'),@((Join-Path $bundle 'OpenConsole.exe'),'b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775'))){if((Get-FileHash $pair[0]).Hash.ToLowerInvariant() -cne $pair[1]){throw 'Pinned binary mismatch'}}
New-Item -ItemType Directory $root|Out-Null
$g=@('--terminal-bundle',$bundle,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-long-fields-run5','--state-dir',(Join-Path $root 'state'))
$driver='asd-long-fields-driver-run5';$data=Join-Path $root 'driver';New-Item -ItemType Directory $data|Out-Null
foreach($entry in @(Get-ChildItem Env:|Where-Object{$_.Name -like 'PSMUX_*' -or $_.Name -in @('TMUX','TMUX_PANE','TMUX_TMPDIR')})){Remove-Item ('Env:'+$entry.Name)}
$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1';$env:PSMUX_DATA_DIR=$data;$env:PSMUX_CONPTY_DIR=$bundle;$env:PSMUX_CONFIG_FILE=Join-Path $root 'psmux.conf'
[IO.File]::WriteAllText($env:PSMUX_CONFIG_FILE,"set -g warm off`nset -g status off`nset -g history-limit 2000`n")
$events=[Collections.Generic.List[object]]::new();$script:longFieldCaptureNumber=0;$s=$null;$target=$null;$manager=$null;$originalShell=$null;$forced=0;$success=$false
function Mux([string[]]$Arguments){$r=(& $mux -L $driver @Arguments|Out-String);if($LASTEXITCODE -ne 0){throw ('Driver failed: '+($Arguments -join ' '))};return $r.TrimEnd("`r","`n")}
function Event($Kind,$Details){$events.Add(@{utc=[DateTime]::UtcNow.ToString('o');kind=$Kind;details=$Details});$events|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'events.json') -Encoding utf8}
function Text([string]$Value){$null=Mux @('send-keys','-t',$target,'-l',$Value);Event 'text' @{value=$Value}}
function Key([string]$Value){$null=Mux @('send-keys','-t',$target,$Value);Event 'key' @{value=$Value}}
function Screen {return Mux @('capture-pane','-t',$target,'-p')}
function Await([string]$Needle){$until=[DateTime]::UtcNow.AddSeconds(10);do{$screen=Screen;if($screen.Contains($Needle)){return $screen};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);[IO.File]::WriteAllText((Join-Path $root 'failed-screen.txt'),$screen);throw ('Screen did not show '+$Needle+'; action not replayed')}
function AwaitClosed {$until=[DateTime]::UtcNow.AddSeconds(10);do{$screen=Screen;if($screen.Contains('管理模式') -and $screen -notmatch '新建终端|重命名会话|本地会话搜索|创建分组|创建子组|重命名分组'){return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);[IO.File]::WriteAllText((Join-Path $root 'failed-screen.txt'),$screen);throw 'Modal close not confirmed; next action refused'}
function CellWidth([string]$Text){$v=$Text.Replace(([char]::ConvertFromUtf32(0x1F469)+[char]0x200D+[char]::ConvertFromUtf32(0x1F4BB)),'xx').Replace(('1'+[char]0xFE0F+[char]0x20E3),'xx').Replace(('e'+[char]0x301),'x');$n=0;foreach($c in $v.ToCharArray()){if([int]$c -ge 0x3400 -and [int]$c -le 0x9fff){$n+=2}else{$n++}};return $n}
function Native([string]$Name){$script:longFieldCaptureNumber++;$screen=Screen;$cursor=(Mux @('display-message','-p','-t',$target,'#{cursor_x}|#{cursor_y}|#{pane_width}|#{pane_height}|#{pane_dead}')).Split('|');$v=@{name=$Name;number=$script:longFieldCaptureNumber;text=$screen;cursor_x=[int]$cursor[0];cursor_y=[int]$cursor[1];width=[int]$cursor[2];height=[int]$cursor[3];pane_dead=$cursor[4]};$v|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root ('capture-'+($script:longFieldCaptureNumber).ToString('000')+'.json')) -Encoding utf8;return $v}
function Field([string]$Name,[string]$Value){Text $Value;$null=Await 'Q';foreach($size in @(@(120,32),@(63,25),@(49,25),@(120,12),@(120,32))){$null=Mux @('resize-window','-t','ui','-x',[string]$size[0],'-y',[string]$size[1]);Event 'resize' @{width=$size[0];height=$size[1]};$null=Await 'Q';Start-Sleep -Milliseconds 160;$v=Native ($Name+'-'+$size[0]+'x'+$size[1]);$rows=$v.text -split "`r?`n";if($v.width -ne $size[0] -or $v.height -ne $size[1] -or $v.cursor_y -ge $rows.Count -or !$rows[$v.cursor_y].Contains('Q')){throw ('Focused native caret not on marker row: '+$v.name)};$at=$rows[$v.cursor_y].IndexOf('Q');if((CellWidth $rows[$v.cursor_y].Substring(0,$at+1)) -ne $v.cursor_x){throw ('Native caret X differs from complete marker: '+$v.name)}};Text '!';$null=Await 'Q!';Key 'BSpace';$null=Await 'Q';Key 'Escape';AwaitClosed}
try{
$fixture=Join-Path $root 'shell.ps1'
@'
function global:prompt { '> ' }
Set-PSReadLineOption -HistorySavePath (Join-Path $PSScriptRoot 'shell.history.txt')
$p=Get-Process -Id $PID
@{pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content (Join-Path $PSScriptRoot 'shell-identity.json') -Encoding utf8
Clear-Host
'LONG_SHELL_READY'
'@|Set-Content $fixture -Encoding utf8
$s=(& $product @g start --name 'Long fields shell' --cwd $root -- $shell -NoLogo -NoProfile -NoExit -File $fixture|Out-String)|ConvertFrom-Json
if($LASTEXITCODE -ne 0){throw 'Shell start failed'};$s|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'session-before.json') -Encoding utf8
$until=[DateTime]::UtcNow.AddSeconds(10);while(!(Test-Path (Join-Path $root 'shell-identity.json')) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 80}
$id=[IO.File]::ReadAllText((Join-Path $root 'shell-identity.json'))|ConvertFrom-Json;$originalShell=Get-Process -Id $id.pid;if($originalShell.Path -ne $shell -or $originalShell.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime){throw 'Original shell mismatch'}
$null=Mux (@('new-session','-d','-s','ui','-c',$root,'-x','120','-y','32','--',$product)+$g)
$pane=(Mux @('list-panes','-t','ui','-F','#{pane_id}')).Trim();if($pane -notmatch '^%\d+$'){throw 'Driver pane mismatch'};$target='ui:.'+$pane
$managerPid=[int](Mux @('display-message','-p','-t',$target,'#{pane_pid}'));$manager=Get-Process -Id $managerPid;if($manager.Path -ne $product){throw 'Manager image mismatch'}
@{pid=$manager.Id;filetime=$manager.StartTime.ToUniversalTime().ToFileTimeUtc();path=$manager.Path}|ConvertTo-Json|Set-Content (Join-Path $root 'manager-identity.json') -Encoding utf8
$registry=Get-ChildItem $data -Filter '*.pid';if(@($registry).Count -ne 1){throw 'Driver registry mismatch'};$parts=([IO.File]::ReadAllText($registry[0].FullName)).Trim().Split(':');$server=Get-Process -Id ([int]$parts[0]);if($server.Path -ne $mux -or $server.StartTime.ToUniversalTime().ToFileTimeUtc() -ne [long]$parts[1]){throw 'Driver identity mismatch'}
$modules=@($server.Modules|Where-Object{$_.FileName -eq (Join-Path $bundle 'conpty.dll')});if($modules.Count -ne 1){throw 'SDK driver module not observed'}
@{pid=$server.Id;filetime=$parts[1];module=$modules[0].FileName;sha256=(Get-FileHash $modules[0].FileName).Hash}|ConvertTo-Json|Set-Content (Join-Path $root 'driver-host.json') -Encoding utf8
$null=Await 'LONG_SHELL_READY';Text "Write-Output 'LONG_DRAFT'";$null=Await "Write-Output 'LONG_DRAFT'";Key 'F6';$null=Await '管理模式'
$value=('中文e'+[char]0x301+[char]::ConvertFromUtf32(0x1F469)+[char]0x200D+[char]::ConvertFromUtf32(0x1F4BB)+'1'+[char]0xFE0F+[char]0x20E3+'/')*6+'ZXtail/tail/tail/Q'
Key 'n';$null=Await '新建终端';Field 'launch-name' $value
Key 'r';$null=Await '重命名会话';Key 'C-u';Field 'session-rename' $value
Key '/';$null=Await '本地会话搜索';Field 'search' $value
Key 'g';$null=Await '创建分组';Field 'group-name' $value
Key 'g';$null=Await '创建分组';Key 'Tab';Field 'group-directory' ('C:\'+$value)
Key '1';Key 'r';$null=Await '重命名分组';Key 'C-u';Field 'group-rename' $value
Key 'F6';$null=Await '终端模式';$null=Await "Write-Output 'LONG_DRAFT'";Key 'Enter';$until=[DateTime]::UtcNow.AddSeconds(10);do{$draftScreen=Screen;if($draftScreen -match '(?m)│\s*LONG_DRAFT\s*│\s*$'){break};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);$null=Native 'draft-output-observation';[IO.File]::WriteAllText((Join-Path $root 'draft-observed.txt'),$draftScreen);if($draftScreen -notmatch '(?m)│\s*LONG_DRAFT\s*│\s*$'){throw 'Explicit draft output missing; Enter not replayed'};$null=Native 'original-draft-explicit-submit'
$after=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or @($after).Count -ne 1 -or $after[0].session_id -ne $s.session_id -or $after[0].run_id -ne $s.run_id){throw 'Original Session/Run changed'};$after|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'session-after.json') -Encoding utf8
if($originalShell.HasExited){throw 'Original shell ended during dialogs'}
Key 'F6';$null=Await '管理模式';Key 'q';if(!$manager.WaitForExit(15000)){throw 'Manager did not exit normally'};if($manager.ExitCode -ne 0){throw 'Manager exit failed'}
$success=$true
}catch{Event 'failure' @{error=$_.Exception.Message};throw}finally{
if($s){& $product @g stop $s.session_id|Set-Content (Join-Path $root 'owned-stop.json') -Encoding utf8;if($LASTEXITCODE -ne 0){Event 'cleanup-error' @{error='Owned stop failed'}}}
if($target){if($manager -and !$manager.HasExited){$forced++};try{$null=Mux @('kill-session','-t','ui')}catch{Event 'cleanup-error' @{error=$_.Exception.Message}}}
if($originalShell){$null=$originalShell.WaitForExit(10000)}
if($server){$null=$server.WaitForExit(10000)}
$result=@{utc=[DateTime]::UtcNow.ToString('o');result=$(if($success){'passed'}else{'failed'});captures=$script:longFieldCaptureNumber;events=$events.Count;forced_product_ends=$forced;manager_ended=($manager -and $manager.HasExited);original_shell_ended=($originalShell -and $originalShell.HasExited);driver_ended=($server -and $server.HasExited);remaining_exact_product=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $product}).Count;clipboard_api_calls=0;desktop_surface='none: isolated psmux/ConPTY driver';windows=[Environment]::OSVersion.VersionString;product_sha256=(Get-FileHash $product).Hash}
$result|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8;$result|ConvertTo-Json -Depth 6
}
