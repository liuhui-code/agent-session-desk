param(
    [Parameter(Mandatory)][string]$PsmuxBinary,
    [Parameter(Mandatory)][string]$PsmuxSHA256,
    [Parameter(Mandatory)][string]$ShellBinary,
    [Parameter(Mandatory)][string]$ArtifactRoot,
    [string]$ConptyDirectory,
    [ValidateSet('hex','paste')][string]$InputMethod='hex',
    [switch]$Utf8Input,
    [switch]$NoResize,
    [switch]$BeforeAttachSeed,
    [switch]$ObserveOnly
)
# Opt-in synthetic native diagnostic. Run in a separate PowerShell process.
# No product host policy, dependency installation or user settings are changed.
$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
. (Join-Path $PSScriptRoot 'windows-process-metrics.ps1')
foreach($path in @($PsmuxBinary,$ShellBinary,$ArtifactRoot)){
    if(![IO.Path]::IsPathRooted($path)){throw 'All paths must be absolute'}
}
if(Test-Path $ArtifactRoot){throw 'Fresh artifact root required'}
$backend=[IO.Path]::GetFullPath($PsmuxBinary);$shell=[IO.Path]::GetFullPath($ShellBinary)
$expectedBackend=$PsmuxSHA256.ToLowerInvariant()
if((Get-FileHash -LiteralPath $backend).Hash.ToLowerInvariant() -cne $expectedBackend){throw 'Backend SHA-256 mismatch'}
$hostFiles=@();$SDK=[bool]$ConptyDirectory;$hostDirectory=$null
if($SDK){
    if(![IO.Path]::IsPathRooted($ConptyDirectory)){throw 'ConPTY directory must be absolute'}
    $hostDirectory=[IO.Path]::GetFullPath($ConptyDirectory)
    # Same explicit measurement candidate as windows-dependencies.json.
    foreach($pair in @(@('conpty.dll','e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492'),@('OpenConsole.exe','b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775'))){
        $path=Join-Path $hostDirectory $pair[0];$item=Get-Item -LiteralPath $path
        if($item.PSIsContainer -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)){throw 'Host files must be regular pinned files'}
        $hash=(Get-FileHash -LiteralPath $path).Hash.ToLowerInvariant();$sig=Get-AuthenticodeSignature -LiteralPath $path
        if($hash -cne $pair[1] -or $sig.Status -ne 'Valid' -or $sig.SignerCertificate.Subject -notmatch '^CN=Microsoft Corporation,'){throw 'SDK hash/signature mismatch'}
        $hostFiles+=,@{path=$path;sha256=$hash;signer=$sig.SignerCertificate.Subject}
    }
}
$root=[IO.Path]::GetFullPath($ArtifactRoot);$Case=Split-Path $root -Leaf
$null=New-Item -ItemType Directory $root
$ns='asd-vt-'+[Guid]::NewGuid().ToString('N').Substring(0,12);$name='vt-'+[Guid]::NewGuid().ToString('N')
$data=Join-Path $root 'runtime';$null=New-Item -ItemType Directory $data
foreach($entry in @(Get-ChildItem Env:|Where-Object{$_.Name -like 'PSMUX_*' -or $_.Name -in @('TMUX','TMUX_PANE','TMUX_TMPDIR')})){Remove-Item ('Env:'+$entry.Name)}
$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1';$env:PSMUX_DATA_DIR=$data
$env:PSMUX_CONFIG_FILE=Join-Path $root 'psmux.conf'
[IO.File]::WriteAllText($env:PSMUX_CONFIG_FILE,"set -g warm off`nset -g status off`nset -g history-limit 2000`n")
if($SDK){$env:PSMUX_CONPTY_DIR=$hostDirectory}
if($BeforeAttachSeed){[IO.File]::WriteAllText((Join-Path $root 'before-attach.txt'),'before')}
if($Utf8Input){[IO.File]::WriteAllText((Join-Path $root 'utf8-input.txt'),'utf8')}
$fixture=Join-Path $root 'fixture.ps1'
@'
$ErrorActionPreference='Stop'
if(Test-Path utf8-input.txt){[Console]::InputEncoding=[Text.UTF8Encoding]::new($false)}
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
if(Test-Path before-attach.txt){Clear-Host;'SCROLL_NEEDLE_中文_α';1..40|ForEach-Object{'history line '+$_};'BEFORE_ATTACH_READY';'seeded'|Set-Content seeded.txt;while(!(Test-Path normal-gate.txt)){Start-Sleep -Milliseconds 20}}
'INPUT_READY'
$text=[Console]::ReadLine();$mode=[uint32]0;$ok=[VtProbeNative]::GetConsoleMode([VtProbeNative]::GetStdHandle(-10),[ref]$mode)
@{text=$text;units=@($text.ToCharArray()|ForEach-Object{[int]$_});input_mode=$mode;mode_available=$ok;input_codepage=[VtProbeNative]::GetConsoleCP()}|ConvertTo-Json|Set-Content normal-input.json -Encoding utf8
while(!(Test-Path seed-gate.txt)){Start-Sleep -Milliseconds 20}
if(!(Test-Path before-attach.txt)){Clear-Host;'SCROLL_NEEDLE_中文_α';1..40|ForEach-Object{'history line '+$_}}
'SEED_READY'
while(!(Test-Path alt-gate.txt)){Start-Sleep -Milliseconds 20}
[Console]::Write("`e[?1049h`e[2J`e[HVT_BOUNDARY_READY 中文 α")
$text=[Console]::ReadLine();$mode=[uint32]0;$ok=[VtProbeNative]::GetConsoleMode([VtProbeNative]::GetStdHandle(-10),[ref]$mode)
@{text=$text;units=@($text.ToCharArray()|ForEach-Object{[int]$_});columns=[Console]::WindowWidth;rows=[Console]::WindowHeight;input_mode=$mode;mode_available=$ok;input_codepage=[VtProbeNative]::GetConsoleCP()}|ConvertTo-Json|Set-Content alt-input.json -Encoding utf8
[Console]::Write("`e[?1049l")
'VT_BOUNDARY_DONE'
'@|Set-Content $fixture -Encoding utf8
$tcp=$null;$identity=$null;$failure=$null;$started=$false;$pane=$null;$frames=[Collections.Generic.List[object]]::new();$captures=[Collections.Generic.List[object]]::new()
$result=@{case=$Case;sdk_requested=[bool]$SDK;input_method=$InputMethod;utf8_input=[bool]$Utf8Input;resize=(!$NoResize);before_attach_seed=[bool]$BeforeAttachSeed;backend_sha256=$expectedBackend;host_files=$hostFiles;namespace=$ns;session=$name;utc=[DateTime]::UtcNow.ToString('o');result='started'}
function Send([string]$line){$b=[Text.Encoding]::UTF8.GetBytes($line+"`n");$stream.Write($b,0,$b.Length);$stream.Flush()}
function Lines($f){return @($f.layout.rows_v2|ForEach-Object{($_.runs.text-join'').TrimEnd()})}
function WaitFrame([string]$label,[scriptblock]$predicate){
 $until=[DateTime]::UtcNow.AddSeconds(12)
 while([DateTime]::UtcNow -lt $until){Send 'dump-state';$line=$reader.ReadLine();if(!$line){throw 'No frame'};if($line -eq 'NC'){continue};$f=$line|ConvertFrom-Json;if($f.session_name -ne $name -or $f.layout.type -ne 'leaf' -or ($null -ne $pane -and $f.layout.id -ne $pane)){throw 'Synthetic Session/pane changed'};if(& $predicate $f){$frames.Add(@{label=$label;columns=$f.layout.cols;rows=$f.layout.rows;backend_alternate_hint=$f.layout.alternate_screen;text=@(Lines $f)});return $f};Start-Sleep -Milliseconds 30}
 throw ('Expected phase not observed: '+$label)
}
function InputLine{
 $hex=([Text.Encoding]::UTF8.GetBytes('ALT输入α')|ForEach-Object{$_.ToString('x2')})-join' '
 if($InputMethod -eq 'paste'){Send ('send-paste '+[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes('ALT输入α'))+" -t %$pane")}else{Send "send-keys -t %$pane -H $hex"};Send "send-keys -t %$pane Enter"
}
function Capture([string]$label){$text=(& $backend -L $ns capture-pane -t ($name+':.%'+$pane) -p -S -200|Out-String);if($LASTEXITCODE -ne 0){throw 'Capture failed'};$ids=@(($text-split'\r?\n')|ForEach-Object{if($_ -match '^history line (\d+)$'){[int]$Matches[1]}});$captures.Add(@{label=$label;text=$text;ids=$ids;complete=(($ids-join',') -ceq ((1..40)-join','))})}
try{
 & $backend -L $ns new-session -d -s $name -c $root -x 120 -y 32 -- $shell -NoLogo -NoProfile -NoExit -File $fixture|Out-Null
 if($LASTEXITCODE -ne 0){throw 'Direct start failed'};$started=$true
 if($BeforeAttachSeed){$until=[DateTime]::UtcNow.AddSeconds(10);while(!(Test-Path (Join-Path $root 'seeded.txt')) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 50};if(!(Test-Path (Join-Path $root 'seeded.txt'))){throw 'Detached seed not observed'}}
 $reg=$ns+'__'+$name;$port=[int]([IO.File]::ReadAllText((Join-Path $data ($reg+'.port')))).Trim();$key=([IO.File]::ReadAllText((Join-Path $data ($reg+'.key')))).Trim()
 $tcp=[Net.Sockets.TcpClient]::new();$tcp.Connect('127.0.0.1',$port);$stream=$tcp.GetStream();$stream.ReadTimeout=15000;$stream.WriteTimeout=3000;$reader=[IO.StreamReader]::new($stream,[Text.UTF8Encoding]::new($false),$false,4096,$true)
 Send "AUTH $key";if($reader.ReadLine() -ne 'OK'){throw 'Authentication failed'};Send 'PERSISTENT';if($BeforeAttachSeed){$f=WaitFrame 'detached-seed' {param($x) ((Lines $x)-join"`n") -match 'BEFORE_ATTACH_READY'};$pane=$f.layout.id;Capture 'detached-before-resize';[IO.File]::WriteAllText((Join-Path $root 'normal-gate.txt'),'go')};Send 'client-attach';Send 'client-size 94 25'
 $f=WaitFrame 'normal-input-ready' {param($x) $x.layout.cols -eq 94 -and $x.layout.rows -eq 25 -and ((Lines $x)-join"`n") -match 'INPUT_READY'};$pane=$f.layout.id
 $identity=[IO.File]::ReadAllText((Join-Path $root 'identity.json'))|ConvertFrom-Json;$result.identity=$identity
 if($SDK){$pidText=([IO.File]::ReadAllText((Join-Path $data ($reg+'.pid')))).Trim();if($pidText -notmatch '^([1-9][0-9]*):([0-9]{17,20})$'){throw 'Unexpected server identity format'};$parts=@($Matches[1],$Matches[2]);$server=Get-Process -Id ([int]$parts[0]);if($server.Path -ne $backend -or $server.StartTime.ToUniversalTime().ToFileTimeUtc() -ne [long]$parts[1]){throw 'Server identity differs'};$mods=@($server.Modules|Where-Object{$_.FileName -eq (Join-Path $hostDirectory 'conpty.dll')}|ForEach-Object{[string]$_.FileName});if($mods.Count -ne 1){throw 'Expected SDK module not observed; fallback rejected'};$result.loaded_sdk_module=$mods[0];$result.server_identity=@{pid=$server.Id;filetime=[long]$parts[1];path=$server.Path}
  $hostUntil=[DateTime]::UtcNow.AddSeconds(5)
  do{$hostSnapshot=@(Get-CimInstance Win32_Process);$hosts=@($hostSnapshot|Where-Object{$_.ParentProcessId -eq $server.Id -and (Get-AsdMetricPathKey $_.ExecutablePath) -ceq (Get-AsdMetricPathKey (Join-Path $hostDirectory 'OpenConsole.exe'))}|ForEach-Object{$hostProcess=Get-Process -Id $_.ProcessId;@{pid=$hostProcess.Id;filetime=$hostProcess.StartTime.ToUniversalTime().ToFileTimeUtc();path=$hostProcess.Path;parent_pid=$server.Id}})
   if($hosts.Count -eq 1){break};Start-Sleep -Milliseconds 100
  }while([DateTime]::UtcNow -lt $hostUntil)
  $result.sdk_host_candidates=@($hostSnapshot|Where-Object{$_.Name -eq 'OpenConsole.exe' -or $_.ParentProcessId -eq $server.Id}|ForEach-Object{@{pid=$_.ProcessId;path=$_.ExecutablePath;parent_pid=$_.ParentProcessId;name=$_.Name}})
  if($hosts.Count -ne 1){throw 'Expected SDK host child not observed; fallback rejected'}
  $result.loaded_sdk_host=$hosts[0]
 }
 InputLine
 $until=[DateTime]::UtcNow.AddSeconds(8);while(!(Test-Path (Join-Path $root 'normal-input.json')) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 50}
 $result.normal_input=[IO.File]::ReadAllText((Join-Path $root 'normal-input.json'))|ConvertFrom-Json
 [IO.File]::WriteAllText((Join-Path $root 'seed-gate.txt'),'go');$null=WaitFrame 'seed-ready' {param($x) ((Lines $x)-join"`n") -match 'SEED_READY'};Capture 'before-alt'
 [IO.File]::WriteAllText((Join-Path $root 'alt-gate.txt'),'go');$null=WaitFrame 'alt-wide' {param($x) ((Lines $x)-join"`n") -match 'VT_BOUNDARY_READY'}
 if(!$NoResize){Send 'client-size 61 18';$null=WaitFrame 'alt-narrow' {param($x) $x.layout.cols -eq 61 -and $x.layout.rows -eq 18 -and ((Lines $x)-join"`n") -match 'VT_BOUNDARY_READY'}}
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
