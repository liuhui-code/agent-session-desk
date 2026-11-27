$round='keys-r10'
try {
$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false);$OutputEncoding=[Console]::OutputEncoding
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$out=Join-Path $base ('stability-ui-'+$round+'-20261127');if(Test-Path $out){throw 'Fresh desktop output required'};New-Item -ItemType Directory $out|Out-Null
$root=Join-Path $base ('stability-ui-product-'+$round+'-20261127');if(Test-Path $root){throw 'Fresh product directory required'};New-Item -ItemType Directory $root|Out-Null
$artifacts=Get-Content -Raw (Join-Path $base 'instance-artifacts-r3.json')|ConvertFrom-Json
foreach($item in $artifacts.PSObject.Properties){if((Get-FileHash (Join-Path $base $item.Name)).Hash.ToLowerInvariant()-cne $item.Value){throw 'Artifact fingerprint changed'}}
$source=if($round.StartsWith('red')){Join-Path $base 'asd-instance-r3.exe'}else{Join-Path $base 'key.exe'}
$expected=if($round.StartsWith('red')){'c7821a3a4c9ec044c51d00d38da47c9d2a4d1d647ae30c7c6c5f068d6b0f4290'}else{"8ef715e3f9cae7d544a8479547ca980c8b846728270b19ed29befbfe371fb432"}
if((Get-FileHash $source).Hash.ToLowerInvariant()-cne $expected){throw 'Fault-test product fingerprint changed'}
Copy-Item $source (Join-Path $root 'asd-instance-r3.exe')
Copy-Item (Join-Path $package 'local-host.json') (Join-Path $root 'local-host.json')
$product=Join-Path $root 'asd-instance-r3.exe';$inspector=Join-Path $base 'instance-r3.test.exe';$fixture=Join-Path $base 'asd-stability-fixture.exe'
if((Get-FileHash $fixture).Hash.ToLowerInvariant()-cne (Get-Content -Raw (Join-Path $base 'stability-artifacts.json')|ConvertFrom-Json).'asd-stability-fixture.exe'){throw 'Fixture changed'}
$hostInfo=Get-Content -Raw (Join-Path $root 'local-host.json')|ConvertFrom-Json
foreach($entry in @($hostInfo.shell,$hostInfo.terminal)){if((Get-FileHash $entry.path).Hash.ToLowerInvariant()-cne $entry.sha256){throw 'Pinned host changed'}}
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
Add-Type @'
using System;using System.Collections.Generic;using System.Runtime.InteropServices;
public class InstanceDesktop {
 [DllImport("user42.dll",CharSet=CharSet.Unicode)]public static extern bool SetProp(IntPtr h,string name,IntPtr data);
 [DllImport("user42.dll",CharSet=CharSet.Unicode)]public static extern IntPtr GetProp(IntPtr h,string name);
 public delegate int PropProc(IntPtr h,IntPtr name,IntPtr data,IntPtr extra);
 [DllImport("user42.dll",CharSet=CharSet.Unicode)]public static extern int EnumPropsEx(IntPtr h,PropProc callback,IntPtr extra);
 [DllImport("user42.dll")]public static extern bool IsWindow(IntPtr h);
 [DllImport("user42.dll")]public static extern bool IsIconic(IntPtr h);
 [DllImport("user42.dll")]public static extern bool ShowWindow(IntPtr h,int command);
 public static string Stamp(IntPtr h){var names=new List<string>();EnumPropsEx(h,(w,n,d,x)=>{if(n.ToInt64()>65535){var s=Marshal.PtrToStringUni(n);if(s.StartsWith("ASD.Attention.Owner.",StringComparison.Ordinal))names.Add(s);}return 1;},IntPtr.Zero);names.Sort(StringComparer.Ordinal);return String.Join(";",names);}
}
'@
$previousDpi=[DeskFlow]::SetThreadDpiAwarenessContext([IntPtr](-4));if($previousDpi-eq [IntPtr]::Zero){throw 'DPI context unavailable'}
$initialHosts=@([DeskFlow]::HostWindows($hostInfo.terminal.path))
$cursor=[DeskFlow+P]::new();$null=[DeskFlow]::GetCursorPos([ref]$cursor)
$r=@{result='started';stages=@();clipboard_calls=0;registry_writes=0;session_content_exported=0;forced_product_ends=0;intentional_owned_ui_crashes=0;automatic_input_replay=$false}
$script:scopes=@();$script:launches=@();$script:managers=@();$script:fixtures=@();$script:windows=@();$script:inspectIndex=0;$script:window=[IntPtr]::Zero;$script:manager=$null
function Stage([string]$name){$script:stage=$name;$r.stages+=@{stage=$name;utc=[DateTime]::UtcNow.ToString('o')}}
function Wait([string]$label,[scriptblock]$check,[int]$seconds=30){$until=[DateTime]::UtcNow.AddSeconds($seconds);do{if(& $check){return};Start-Sleep -Milliseconds 150}while([DateTime]::UtcNow-lt $until);throw ('Timed out: '+$label+'; no automatic replay')}
function Product($scope,[string[]]$arguments){$g=$scope.globals;$text=(& $product @g @arguments 2>$null|Out-String);if($LASTEXITCODE-ne 0){throw 'Controlled product CLI failed'};return $text}
function Journal([string]$path){
 if(!(Test-Path $path)){return @()};$f=[IO.File]::Open($path,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::ReadWrite-bor[IO.FileShare]::Delete)
 try{$reader=[IO.StreamReader]::new($f);try{$text=$reader.ReadToEnd()}finally{$reader.Dispose()}}finally{$f.Dispose()}
 $end=$text.LastIndexOf("`n");if($end-lt 0){return @()};return @($text.Substring(0,$end)-split "`r?`n"|Where-Object{$_}|ForEach-Object{$_|ConvertFrom-Json})
}
function Inspect($scope){
 $script:inspectIndex++;if($script:inspectIndex-gt 240){throw 'Observer bound exceeded'}
 $file=Join-Path $out ('inspect-'+$script:inspectIndex+'.json')
 $env:ASD_INSTANCE_INSPECT_OUTPUT=$file;$env:ASD_INSTANCE_INSPECT_LOCATION=Join-Path $scope.runtime $scope.ns;$env:ASD_INSTANCE_INSPECT_METADATA=Join-Path $scope.state 'metadata.sqlite3'
 & $inspector '-test.run' '^TestInstanceInspectHelper$' '-test.timeout' '10s' 1>$null 2>$null
 if($LASTEXITCODE-ne 0 -or !(Test-Path $file)){throw 'Read-only synthetic scope observer failed'}
 return Get-Content -Raw $file|ConvertFrom-Json
}
function Launch($scope){
 $info=[Diagnostics.ProcessStartInfo]::new();$info.FileName=$product;$info.UseShellExecute=$false;$info.WorkingDirectory=$scope.cwd
 foreach($arg in @($scope.globals)+@('--pet','--document-app','system-notepad')){$info.ArgumentList.Add($arg)}
 $p=[Diagnostics.Process]::Start($info);$null=$p.Handle;$script:launches+=$p;return $p
}
function WaitLauncher($p){if(!$p.WaitForExit(25000)-or $p.ExitCode-ne 0){throw 'Controlled launcher failed or timed out'}}
function RegisterWindow($scope){
 Wait 'one bound synthetic scope' {
  $id=Inspect $scope;if($id.owner_pid-eq 0 -or $id.target-eq 0){return $false}
  $w=[IntPtr]([long]$id.target);if($w-in $initialHosts){throw 'Endpoint points at a pre-existing user window; no UI actions'}
  foreach($other in $script:scopes){if($other-ne $scope -and $other.ContainsKey('window') -and $other.window-eq $w){throw 'Scopes point at the same native window'}}
  $p=Get-Process -Id $id.owner_pid -ErrorAction SilentlyContinue;if(!$p -or $p.Path-ine $product){throw 'Endpoint owner path mismatch'};$null=$p.Handle
  if([DeskFlow]::Title($w)-cne 'Agent Session Desk' -or [DeskFlow]::OwnerBindings($w)-ne 1){return $false}
  $pets=@([DeskFlow]::Find('ASD Blocking Reminders',$p.Id));if($pets.Count-ne 1){return $false}
  $scope.manager=$p;$scope.window=$w;$scope.pet=$pets[0];$scope.stamp="driver-bound"
  $script:manager=$p;$script:window=$w;$script:managers+=$p
  $script:windows+=@{scope=$scope;handle=$w;stamp=$scope.stamp;manager=$p}
  return $true
 }
}
function Surface($scope){
 if(!$scope.ContainsKey('driver_stamp')-or ![DeskFlow]::IsWindowVisible($scope.window)-or [DeskFlow]::OwnerBindings($scope.window)-ne 1-or $scope.host.HasExited){throw 'Driver-owned window identity changed'}
 $rootElement=[Windows.Automation.AutomationElement]::FromHandle($scope.window)
 $items=$rootElement.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition);$best=$null
 foreach($item in $items){$pattern=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){$text=$pattern.DocumentRange.GetText(65536);if(!$best-or $text.Length-gt $best.text.Length){$best=@{text=$text;pattern=$pattern}}}}
 if(!$best){throw 'Driver-owned fixture surface unavailable'};return $best
}
function Ready($scope,[string]$marker){Wait 'original fixture surface ready' {$v=Surface $scope;return $v.text.Contains($marker)-and !$v.text.Contains('Connecting to')-and !$v.text.Contains('Read-only last')-and !$v.text.Contains('Terminal not interactive')}}
function Guard($scope){
 if($scope.manager.HasExited-or [DeskFlow]::GetForegroundWindow()-ne $scope.window){throw 'Owned foreground changed; no input replay'}
 foreach($key in @(1,2,4,16,17,18,91,92)){if(([DeskFlow]::GetAsyncKeyState($key)-band 32768)-ne 0){throw 'User key/button held; no input'}}
 $null=Surface $scope
}
function ClickAt([double]$x,[double]$y){$script:lastX=[int]$x;$script:lastY=[int]$y;if([DeskFlow]::Move($script:lastX,$script:lastY)-ne 1 -or [DeskFlow]::Click()-ne 2){throw 'Owned click incomplete'}}
function ClickText($scope,[string]$text){$v=Surface $scope;$range=$v.pattern.DocumentRange.FindText($text,$false,$false);if(!$range){throw 'Controlled text range missing'};$rects=@($range.GetBoundingRectangles());if($rects.Count-ne 1){throw 'Controlled text range clipped'};Guard $scope;ClickAt ($rects[0].Left+$rects[0].Width/2) ($rects[0].Top+$rects[0].Height/2)}
function Foreground($scope){$null=[DeskFlow]::SetForegroundWindow($scope.window);Wait 'owned fixture foreground' {[DeskFlow]::GetForegroundWindow()-eq $scope.window};Guard $scope}
function CreateScope([string]$label){
 $cwd=Join-Path $out $label;New-Item -ItemType Directory $cwd|Out-Null
 $runtime=Join-Path $cwd 'runtime';$state=Join-Path $cwd 'state';$ns='asd-inst-'+$label+'-20261126'
 $scope=@{label=$label;cwd=$cwd;runtime=$runtime;state=$state;ns=$ns;log=(Join-Path $cwd 'fixture.jsonl')}
 $scope.globals=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',$runtime,'--namespace',$ns,'--state-dir',$state)
 $script:scopes+=$scope
 $record=(Product $scope @('start','--name',('Instance '+$label),'--cwd',$cwd,'--env',('ASD_FIXTURE_REPORT_PATH='+$scope.log),'--',$fixture))|ConvertFrom-Json
 $scope.original=$record
 Wait 'fixture identity receipt' {@(Journal $scope.log).Count-gt 0}
 $start=@(Journal $scope.log)[0];$p=Get-Process -Id $start.pid;$null=$p.Handle;if($p.Path-ine $fixture){throw 'Original fixture path mismatch'};$scope.fixture=$p;$script:fixtures+=$p
 return $scope
}
function AssertOriginal($scope){
 if($scope.fixture.HasExited){throw 'Original fixture ended'}
 if(@(Journal $scope.log|Where-Object{$_.event-eq 'input'}).Count){throw 'Original draft was submitted'}
 $rows=(Product $scope @('list','--json'))|ConvertFrom-Json;$original=$scope.original
 if(@($rows|Where-Object{$_.session_id-ceq $original.session_id-and $_.run_id-ceq $original.run_id-and $_.backend_id-ceq $original.backend_id}).Count-ne 1){throw 'Original Session/Run/pane changed'}
}
function Quit($scope){
 Foreground $scope;ClickText $scope ('Instance '+$scope.label)
 Wait 'management footer' {(Surface $scope).text.Contains('Alt+J/K')}
 Guard $scope;if([DeskFlow]::Text('q')-ne 2){throw 'Management quit incomplete'}
 if(!$scope.manager.WaitForExit(15000)-or $scope.manager.ExitCode-ne 0){throw 'Owned manager did not quit normally'}
 Wait 'instance endpoint gone' {$id=Inspect $scope;return $id.endpoint-eq 0}
 Wait 'owned window ended' {return ![DeskFlow]::IsWindowVisible($scope.window)}
 AssertOriginal $scope
}
function CloseOwnedWindow($entry){
 $w=$entry.handle;if(![DeskFlow]::IsWindowVisible($w)){return}
 if(!$entry.scope.ContainsKey('driver_stamp')-or [DeskFlow]::OwnerBindings($w)-ne 1-or $entry.scope.host.HasExited){throw 'Driver-owned window stamp changed; no close'}
 $rootElement=[Windows.Automation.AutomationElement]::FromHandle($w)
 $tabs=$rootElement.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem))
 if($tabs.Count-gt 1){throw 'Owned window gained another native tab; no close'}
 $null=[DeskFlow]::SendMessage($w,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)
}





$script:originals=@();$a=$null
function Press($k){Guard $a;$null=[DeskFlow]::Key($k)}
function Chord($mod,$k){Guard $a;$null=[DeskFlow]::Chord($mod,$k)}
function Bind {RegisterWindow $a;$a.driver_stamp='ASD.Keys.'+[Guid]::NewGuid();[uint32]$hp=0;$null=[DeskFlow]::GetWindowThreadProcessId($a.window,[ref]$hp);$a.host=Get-Process -Id $hp;$null=$a.host.Handle;$a.bound=$true}
try{
Stage 'configure'
$a=CreateScope 'A';$script:originals+=$a.original
@'
schema_version=1
[global]
wake="ctrl+alt+f9"
[workspace]
next="alt+x"
previous="alt+z"
[management]
open_project="f8"
'@|Set-Content (Join-Path $a.state 'keybindings.toml')
$recipe="schema_version=1`n[templates.one]`nname='one'`nexecutable='"+$fixture+"'`ncwd='"+$a.cwd+"'`n[templates.two]`nname='two'`nexecutable='"+$fixture+"'`ncwd='"+$a.cwd+"'`n[workspaces.default]`ntemplates=['one','two']"
$recipe|Set-Content (Join-Path $a.state 'launch.toml')
Stage 'open'
$p=Launch $a;WaitLauncher $p;Bind;Ready $a 'CLI fixture';Foreground $a
Stage 'workspace'
ClickText $a 'Instance A';Wait 'management' {(Surface $a).text.Contains('Navigation')};Press 0x77;Start-Sleep -Seconds 2
Wait '3 live' {$rows=@((Product $a @('list','--json'))|ConvertFrom-Json);if($rows.Count-eq 3){$script:originals=$rows;return $true};return $false}
Wait 'project result' {(Surface $a).text.Contains('one')-and (Surface $a).text.Contains('two')}
Stage 'custom switching'
ClickText $a 'one';Wait 'management' {(Surface $a).text.Contains('Navigation')};Press 0x0d;Ready $a 'CLI fixture';ClickText $a 'Instance A';Wait 'management' {(Surface $a).text.Contains('Navigation')};Press 0x0d;Ready $a 'CLI fixture';Chord 0x12 0x58;Wait 'one active' {(Surface $a).text.Contains('┃ one')};Start-Sleep -Seconds 1;Chord 0x12 0x5a;Wait 'A active' {(Surface $a).text.Contains('┃ Instance A')};Ready $a 'CLI fixture'
Stage 'global restore'
$null=[DeskFlow]::SendMessage($a.window,0x112,[UIntPtr]0xf020,[IntPtr]::Zero);Start-Sleep -Milliseconds 250
$null=[DeskFlow]::Keys([ushort[]]@(0x11,0x12),[uint[]]@(0,0));$null=[DeskFlow]::Key(0x78);$null=[DeskFlow]::Keys([ushort[]]@(0x12,0x11),[uint[]]@(2,2))
Wait 'foreground restored' {[DeskFlow]::GetForegroundWindow()-eq $a.window-and !!([DeskFlow]::IsWindowVisible($a.window))}
Stage 'quit'
ClickText $a 'Instance A';Wait 'management' {(Surface $a).text.Contains('Navigation')};Guard $a;$null=[DeskFlow]::Text('q');if(!$a.manager.WaitForExit(10000)){throw 'No normal quit'}
$r.result='verified';$r.project_live=3;$r.global_restored=$true
}catch{$r.result='failed';$r.error=$_.Exception.Message;$r.stage=$script:stage}
finally{
foreach($e in $script:windows){try{CloseOwnedWindow $e}catch{}}
foreach($v in $script:originals){try{$null=Product $a @('stop',$v.session_id)}catch{}}
$r|ConvertTo-Json -Depth 5|Set-Content (Join-Path $out 'result')
}
}catch{$_|Out-File $env:TEMP/key-error}
