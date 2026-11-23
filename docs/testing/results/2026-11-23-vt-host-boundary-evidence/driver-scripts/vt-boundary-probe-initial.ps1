param([string]$Case='candidate-system1',[switch]$SDK,[switch]$Official,[switch]$ObserveOnly)
$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base ('vt-boundary-'+$Case)
if(Test-Path $root){throw 'Fresh root required'};New-Item -ItemType Directory $root|Out-Null
$backend=if($Official){Join-Path $base 'backend/psmux.exe'}else{Join-Path $base 'history-backend-candidate/psmux.exe'}
$expectedBackend=if($Official){'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'}else{'ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'}
if((Get-FileHash $backend).Hash.ToLowerInvariant() -ne $expectedBackend){throw 'Backend hash mismatch'}
$shell=Join-Path $base 'pwsh7/pwsh.exe';$ns='asd-vt-'+[Guid]::NewGuid().ToString('N').Substring(0,12);$name='vt-'+[Guid]::NewGuid().ToString('N');$data=Join-Path $root 'runtime';New-Item -ItemType Directory $data|Out-Null
foreach($entry in @(Get-ChildItem Env:|Where-Object{$_.Name -like 'PSMUX_*' -or $_.Name -in @('TMUX','TMUX_PANE','TMUX_TMPDIR')})){Remove-Item ('Env:'+$entry.Name)}
$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1';$env:PSMUX_DATA_DIR=$data;$env:PSMUX_CONFIG_FILE=Join-Path $root 'psmux.conf';[IO.File]::WriteAllText($env:PSMUX_CONFIG_FILE,"set -g warm off`nset -g status off`nset -g history-limit 2000`n")
$hostFiles=@();$hostDirectory=Join-Path $base 'terminal/terminal-1.25.3131.0'
if($SDK){
 foreach($pair in @(@('conpty.dll','e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492'),@('OpenConsole.exe','b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775'))){$p=Join-Path $hostDirectory $pair[0];$h=(Get-FileHash $p).Hash.ToLowerInvariant();$sig=Get-AuthenticodeSignature $p;if($h -ne $pair[1] -or $sig.Status -ne 'Valid' -or $sig.SignerCertificate.Subject -notmatch '^CN=Microsoft Corporation,'){throw 'SDK hash/signature mismatch'};$hostFiles+=,@{path=$p;sha256=$h;signer=$sig.SignerCertificate.Subject}}
 $env:PSMUX_CONPTY_DIR=$hostDirectory
}
$fixture=Join-Path $root 'fixture.ps1'
@'
$ErrorActionPreference='Stop'
Add-Type @"
using System;using System.Runtime.InteropServices;
public class VtProbeNative {
 [DllImport("kernel32.dll")] public static extern IntPtr GetStdHandle(int n);
 [DllImport("kernel32.dll")] public static extern bool GetConsoleMode(IntPtr h,out uint m);
 [DllImport("kernel32.dll")] public static extern uint GetConsoleCP();
}
"@
$p=Get-Process -Id $PID
@{pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content identity.json -Encoding utf8
'INPUT_READY'
$text=[Console]::ReadLine();$mode=[uint32]0;$ok=[VtProbeNative]::GetConsoleMode([VtProbeNative]::GetStdHandle(-10),[ref]$mode)
@{text=$text;units=@($text.ToCharArray()|ForEach-Object{[int]$_});input_mode=$mode;mode_available=$ok;input_codepage=[VtProbeNative]::GetConsoleCP()}|ConvertTo-Json|Set-Content normal-input.json -Encoding utf8
while(!(Test-Path seed-gate.txt)){Start-Sleep -Milliseconds 20}
Clear-Host
'SCROLL_NEEDLE_中文_α'
1..40|ForEach-Object{'history line '+$_}
'SEED_READY'
while(!(Test-Path alt-gate.txt)){Start-Sleep -Milliseconds 20}
[Console]::Write("`e[?1049h`e[2J`e[HVT_BOUNDARY_READY 中文 α")
$text=[Console]::ReadLine();$mode=[uint32]0;$ok=[VtProbeNative]::GetConsoleMode([VtProbeNative]::GetStdHandle(-10),[ref]$mode)
@{text=$text;units=@($text.ToCharArray()|ForEach-Object{[int]$_});columns=[Console]::WindowWidth;rows=[Console]::WindowHeight;input_mode=$mode;mode_available=$ok;input_codepage=[VtProbeNative]::GetConsoleCP()}|ConvertTo-Json|Set-Content alt-input.json -Encoding utf8
[Console]::Write("`e[?1049l")
'VT_BOUNDARY_DONE'
'@|Set-Content $fixture -Encoding utf8
$tcp=$null;$identity=$null;$failure=$null;$started=$false;$pane=$null;$frames=[Collections.Generic.List[object]]::new();$captures=[Collections.Generic.List[object]]::new()
$result=@{case=$Case;official=[bool]$Official;sdk_requested=[bool]$SDK;backend_sha256=$expectedBackend;host_files=$hostFiles;namespace=$ns;session=$name;utc=[DateTime]::UtcNow.ToString('o');result='started'}
function Send([string]$line){$b=[Text.Encoding]::UTF8.GetBytes($line+"`n");$stream.Write($b,0,$b.Length);$stream.Flush()}
function Lines($f){return @($f.layout.rows_v2|ForEach-Object{($_.runs.text-join'').TrimEnd()})}
function WaitFrame([string]$label,[scriptblock]$predicate){
 $until=[DateTime]::UtcNow.AddSeconds(12)
 while([DateTime]::UtcNow -lt $until){Send 'dump-state';$line=$reader.ReadLine();if(!$line){throw 'No frame'};if($line -eq 'NC'){continue};$f=$line|ConvertFrom-Json;if($f.session_name -ne $name -or $f.layout.type -ne 'leaf' -or ($null -ne $pane -and $f.layout.id -ne $pane)){throw 'Synthetic Session/pane changed'};if(& $predicate $f){$frames.Add(@{label=$label;columns=$f.layout.cols;rows=$f.layout.rows;backend_alternate_hint=$f.layout.alternate_screen;text=@(Lines $f)});return $f};Start-Sleep -Milliseconds 30}
 throw ('Expected phase not observed: '+$label)
}
function InputLine{
 $hex=([Text.Encoding]::UTF8.GetBytes('ALT输入α')|ForEach-Object{$_.ToString('x2')})-join' '
 Send "send-keys -t %$pane -H $hex";Send "send-keys -t %$pane Enter"
}
function Capture([string]$label){$text=(& $backend -L $ns capture-pane -t ($name+':.%'+$pane) -p -S -200|Out-String);if($LASTEXITCODE -ne 0){throw 'Capture failed'};$ids=@(($text-split'\r?\n')|ForEach-Object{if($_ -match '^history line (\d+)$'){[int]$Matches[1]}});$captures.Add(@{label=$label;text=$text;ids=$ids;complete=(($ids-join',') -ceq ((1..40)-join','))})}
try{
 & $backend -L $ns new-session -d -s $name -c $root -x 120 -y 32 -- $shell -NoLogo -NoProfile -NoExit -File $fixture|Out-Null
 if($LASTEXITCODE -ne 0){throw 'Direct start failed'};$started=$true
 $reg=$ns+'__'+$name;$port=[int]([IO.File]::ReadAllText((Join-Path $data ($reg+'.port')))).Trim();$key=([IO.File]::ReadAllText((Join-Path $data ($reg+'.key')))).Trim()
 $tcp=[Net.Sockets.TcpClient]::new();$tcp.Connect('127.0.0.1',$port);$stream=$tcp.GetStream();$stream.ReadTimeout=15000;$stream.WriteTimeout=3000;$reader=[IO.StreamReader]::new($stream,[Text.UTF8Encoding]::new($false),$false,4096,$true)
 Send "AUTH $key";if($reader.ReadLine() -ne 'OK'){throw 'Authentication failed'};Send 'PERSISTENT';Send 'client-attach';Send 'client-size 94 25'
 $f=WaitFrame 'normal-input-ready' {param($x) $x.layout.cols -eq 94 -and $x.layout.rows -eq 25 -and ((Lines $x)-join"`n") -match 'INPUT_READY'};$pane=$f.layout.id
 $identity=[IO.File]::ReadAllText((Join-Path $root 'identity.json'))|ConvertFrom-Json;$result.identity=$identity
 if($SDK){$parts=([IO.File]::ReadAllText((Join-Path $data ($reg+'.pid')))).Trim()-split'\s+';$server=Get-Process -Id ([int]$parts[0]);if($server.Path -ne $backend -or $server.StartTime.ToUniversalTime().ToFileTimeUtc() -ne [long]$parts[1]){throw 'Server identity differs'};$mods=@($server.Modules|Where-Object{$_.FileName -eq (Join-Path $hostDirectory 'conpty.dll')}|ForEach-Object{[string]$_.FileName});if($mods.Count -ne 1){throw 'Expected SDK module not observed; fallback rejected'};$result.loaded_sdk_module=$mods[0]}
 InputLine
 $until=[DateTime]::UtcNow.AddSeconds(8);while(!(Test-Path (Join-Path $root 'normal-input.json')) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 50}
 $result.normal_input=[IO.File]::ReadAllText((Join-Path $root 'normal-input.json'))|ConvertFrom-Json
 [IO.File]::WriteAllText((Join-Path $root 'seed-gate.txt'),'go');$null=WaitFrame 'seed-ready' {param($x) ((Lines $x)-join"`n") -match 'SEED_READY'};Capture 'before-alt'
 [IO.File]::WriteAllText((Join-Path $root 'alt-gate.txt'),'go');$null=WaitFrame 'alt-wide' {param($x) ((Lines $x)-join"`n") -match 'VT_BOUNDARY_READY'}
 Send 'client-size 61 18';$null=WaitFrame 'alt-narrow' {param($x) $x.layout.cols -eq 61 -and $x.layout.rows -eq 18 -and ((Lines $x)-join"`n") -match 'VT_BOUNDARY_READY'}
 InputLine;$null=WaitFrame 'after-alt' {param($x) ((Lines $x)-join"`n") -match 'VT_BOUNDARY_DONE'};Capture 'after-alt'
 $result.alt_input=[IO.File]::ReadAllText((Join-Path $root 'alt-input.json'))|ConvertFrom-Json
 $result.normal_input_exact=($result.normal_input.text -ceq 'ALT输入α');$result.alt_input_exact=($result.alt_input.text -ceq 'ALT输入α');$result.history_complete=(@($captures|Where-Object{!$_.complete}).Count -eq 0)
 $result.result=if($result.normal_input_exact -and $result.alt_input_exact -and $result.history_complete){'complete-in-probe'}else{'observed-incomplete'}
}catch{$failure=$_.Exception.Message;$result.result='harness-failed';$result.error=$failure}
finally{
 if($tcp){$tcp.Dispose()};if($started){& $backend -L $ns kill-session -t $name|Out-Null;if($LASTEXITCODE -ne 0){$result.cleanup_error='Stop failed'}}
 if($identity){$p=Get-Process -Id $identity.pid -ErrorAction SilentlyContinue;$result.original_ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $identity.filetime -or $p.Path -ne $identity.path)}
 $frames.ToArray()|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'frames.json') -Encoding utf8;$captures.ToArray()|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'captures.json') -Encoding utf8;$result|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'result.json') -Encoding utf8
}
$result|ConvertTo-Json -Depth 10
if($failure -or $result.cleanup_error -or !$result.original_ended){throw 'Probe or cleanup failed'}
if(!$ObserveOnly -and $result.result -ne 'complete-in-probe'){throw 'VT/input boundary requirements failed'}
