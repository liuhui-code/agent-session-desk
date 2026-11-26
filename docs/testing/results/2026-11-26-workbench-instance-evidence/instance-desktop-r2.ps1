try {
$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$out=Join-Path $base 'instance-desktop-r2-20261126';if(Test-Path $out){throw 'Fresh desktop output required'};New-Item -ItemType Directory $out|Out-Null
$root=Join-Path $base 'instance-final-product-r2-20261126';if(Test-Path $root){throw 'Fresh product directory required'};New-Item -ItemType Directory $root|Out-Null
$artifacts=Get-Content -Raw (Join-Path $base 'instance-artifacts-r3.json')|ConvertFrom-Json
foreach($item in $artifacts.PSObject.Properties){if((Get-FileHash (Join-Path $base $item.Name)).Hash.ToLowerInvariant()-cne $item.Value){throw 'Artifact fingerprint changed'}}
Copy-Item (Join-Path $base 'asd-instance-r3.exe') (Join-Path $root 'asd-instance-r3.exe')
Copy-Item (Join-Path $package 'local-host.json') (Join-Path $root 'local-host.json')
$product=Join-Path $root 'asd-instance-r3.exe';$inspector=Join-Path $base 'instance-r3.test.exe';$fixture=Join-Path $base 'asd-fixture-integrated.exe'
if((Get-FileHash $fixture).Hash.ToLowerInvariant()-cne 'b4405f298f3beef8cf61f4fff12d7d71f37f74d5e323535dd1076616f477ede3'){throw 'Fixture changed'}
$hostInfo=Get-Content -Raw (Join-Path $root 'local-host.json')|ConvertFrom-Json
foreach($entry in @($hostInfo.shell,$hostInfo.terminal)){if((Get-FileHash $entry.path).Hash.ToLowerInvariant()-cne $entry.sha256){throw 'Pinned host changed'}}
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
Add-Type @'
using System;using System.Collections.Generic;using System.Runtime.InteropServices;
public class InstanceDesktop {
 [DllImport("user32.dll")]public static extern bool IsWindow(IntPtr h);
 [DllImport("user32.dll")]public static extern bool IsIconic(IntPtr h);
 [DllImport("user32.dll")]public static extern bool ShowWindow(IntPtr h,int command);
 public static string Stamp(IntPtr h){var names=new List<string>();DeskFlow.EnumPropsEx(h,(w,n,d,x)=>{if(n.ToInt64()>65535){var s=Marshal.PtrToStringUni(n);if(s.StartsWith("ASD.Attention.Owner.",StringComparison.Ordinal))names.Add(s);}return 1;},IntPtr.Zero);names.Sort(StringComparer.Ordinal);return String.Join(";",names);}
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
 $script:inspectIndex++;if($script:inspectIndex-gt 160){throw 'Observer bound exceeded'}
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
  $scope.manager=$p;$scope.window=$w;$scope.pet=$pets[0];$scope.stamp=[InstanceDesktop]::Stamp($w)
  $script:manager=$p;$script:window=$w;$script:managers+=$p
  $script:windows+=@{scope=$scope;handle=$w;stamp=$scope.stamp;manager=$p}
  return $true
 }
}
function Surface($scope){
 if(!$scope.ContainsKey('window')-or ![InstanceDesktop]::IsWindow($scope.window)-or [InstanceDesktop]::Stamp($scope.window)-cne $scope.stamp){throw 'Owned native window identity changed'}
 $rootElement=[Windows.Automation.AutomationElement]::FromHandle($scope.window)
 $items=$rootElement.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition);$best=$null
 foreach($item in $items){$pattern=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){$text=$pattern.DocumentRange.GetText(65536);if(!$best-or $text.Length-gt $best.text.Length){$best=@{text=$text;pattern=$pattern}}}}
 if(!$best){throw 'Owned fixture surface unavailable'};return $best
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
 Wait 'owned window ended' {return ![InstanceDesktop]::IsWindow($scope.window)}
 AssertOriginal $scope
}
function CloseOwnedWindow($entry){
 $w=$entry.handle;if(![InstanceDesktop]::IsWindow($w)){return}
 if([InstanceDesktop]::Stamp($w)-cne $entry.stamp){if($entry.manager.HasExited){$r.reused_or_closed_window_skipped=$true;return};throw 'Owned window stamp changed; no close'}
 $rootElement=[Windows.Automation.AutomationElement]::FromHandle($w)
 $tabs=$rootElement.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem))
 if($tabs.Count-gt 1){throw 'Owned window gained another native tab; no close'}
 $null=[DeskFlow]::SendMessage($w,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)
}
try{
 Stage 'input-idle-preflight'
 Wait 'user input idle' {[DeskFlow]::IdleMilliseconds()-ge 45000} 60
 $a=CreateScope 'A';$b=CreateScope 'B'
 Stage 'five-simultaneous-first-launches'
 $initial=@();for($i=0;$i-lt 5;$i++){$initial+=Launch $a};foreach($p in $initial){WaitLauncher $p}
 RegisterWindow $a;Ready $a 'CLI fixture'
 $r.first_five_launches_one_owner=$true
 Foreground $a;ClickText $a 'CLI fixture';Guard $a;if([DeskFlow]::Text('DRAFT_INSTANCE_A')-ne 32){throw 'Draft A injection incomplete'};Ready $a 'DRAFT_INSTANCE_A'
 Stage 'second-scope-same-terminal-broker'
 $p=Launch $b;WaitLauncher $p;RegisterWindow $b;Ready $b 'CLI fixture'
 Foreground $b;ClickText $b 'CLI fixture';Guard $b;if([DeskFlow]::Text('DRAFT_INSTANCE_B')-ne 32){throw 'Draft B injection incomplete'};Ready $b 'DRAFT_INSTANCE_B'
 [uint32]$ap=0;[uint32]$bp=0;$null=[DeskFlow]::GetWindowThreadProcessId($a.window,[ref]$ap);$null=[DeskFlow]::GetWindowThreadProcessId($b.window,[ref]$bp)
 $r.same_terminal_broker=($ap-eq $bp);$r.separate_scope_windows=($a.window-ne $b.window)
 Stage 'pet-a-returns-from-b'
 # Hide only this test's B pet to expose A at the shared initial position.
 $null=[DeskFlow]::SendMessage($b.pet,0x111,[UIntPtr]5,[IntPtr]::Zero)
 $rect=[DeskFlow+R]::new();$null=[DeskFlow]::GetWindowRect($a.pet,[ref]$rect);Guard $b
 ClickAt (($rect.Left+$rect.Right)/2) ($rect.Top+($rect.Bottom-$rect.Top)*0.92)
 Wait 'A pet restores A, not foreground B' {[DeskFlow]::GetForegroundWindow()-eq $a.window};Ready $a 'DRAFT_INSTANCE_A';AssertOriginal $a;AssertOriginal $b
 $r.pet_restored_correct_scope_from_other_window=$true
 Stage 'five-duplicate-launches-preserve-original-owner'
 $originalPID=$a.manager.Id
 for($i=0;$i-lt 5;$i++){Foreground $b;$p=Launch $a;WaitLauncher $p;Wait 'duplicate launch restores A' {[DeskFlow]::GetForegroundWindow()-eq $a.window};$id=Inspect $a;if($id.owner_pid-ne $originalPID-or [IntPtr]([long]$id.target)-ne $a.window){throw 'Duplicate launch replaced original owner'}}
 $r.repeated_launch_kept_original_owner=$true
 Stage 'duplicate-launch-restores-minimized-a'
 Guard $a;$null=[InstanceDesktop]::ShowWindow($a.window,6);Wait 'A minimized' {[InstanceDesktop]::IsIconic($a.window)}
 Foreground $b;$p=Launch $a;WaitLauncher $p
 Wait 'duplicate restores minimized A' {[DeskFlow]::GetForegroundWindow()-eq $a.window-and ![InstanceDesktop]::IsIconic($a.window)}
 $r.minimized_duplicate_restored_original=$true
 AssertOriginal $a;AssertOriginal $b
 Stage 'normal-quit-and-reopen-original-run'
 Quit $a;$p=Launch $a;WaitLauncher $p;RegisterWindow $a;Ready $a 'DRAFT_INSTANCE_A';AssertOriginal $a
 $r.normal_quit_released_instance_and_preserved_run=$true
 Stage 'intentional-owned-ui-crash-and-reopen'
 $crashed=$a.manager;$entry=$script:windows[-1];$crashed.Kill();if(!$crashed.WaitForExit(10000)){throw 'Owned crash fixture did not end'};$r.intentional_owned_ui_crashes=1
 Wait 'crashed instance endpoint gone' {$id=Inspect $a;return $id.endpoint-eq 0}
 if(@([DeskFlow]::Find('ASD Blocking Reminders',$crashed.Id)).Count){throw 'Crashed UI left a live pet'}
 CloseOwnedWindow $entry;Wait 'crashed owned host ended' {return ![InstanceDesktop]::IsWindow($entry.handle)}
 AssertOriginal $a;$p=Launch $a;WaitLauncher $p;RegisterWindow $a;Ready $a 'DRAFT_INSTANCE_A';AssertOriginal $a;AssertOriginal $b
 $r.crash_released_instance_and_preserved_run=$true
 Stage 'final-normal-quit'
 Quit $a;Quit $b
 $r.result='verified';$r.original_runs_and_unsubmitted_drafts_preserved=$true;$r.no_fixture_input_submitted=$true
}catch{$r.result='failed';$r.stage=$script:stage;$r.error=$_.Exception.Message.Replace($env:USERPROFILE,'<USERPROFILE>')}
finally{
 foreach($entry in $script:windows){try{CloseOwnedWindow $entry}catch{$r.owned_window_cleanup_blocked=$true}}
 foreach($scope in $script:scopes){if($scope.ContainsKey('original')){try{$rows=(Product $scope @('list','--json'))|ConvertFrom-Json;$o=$scope.original;if(@($rows|Where-Object{$_.session_id-ceq $o.session_id-and $_.run_id-ceq $o.run_id-and $_.backend_id-ceq $o.backend_id}).Count-eq 1){$null=Product $scope @('stop',$o.session_id)}else{$r.fixture_cleanup_identity_changed=$true}}catch{$r.fixture_cleanup_error=$true}}}
 $r.fixture_ends=@($script:fixtures|ForEach-Object{if(!$_.HasExited){$null=$_.WaitForExit(10000)};$_.HasExited})
 $r.manager_ends=@($script:managers|ForEach-Object{if(!$_.HasExited){$null=$_.WaitForExit(10000)};$_.HasExited})
 $r.launcher_ends=@($script:launches|ForEach-Object{if(!$_.HasExited){$null=$_.WaitForExit(10000)};$_.HasExited})
 $now=[DeskFlow+P]::new();$null=[DeskFlow]::GetCursorPos([ref]$now);if($null-ne $script:lastX-and $now.X-eq $script:lastX-and $now.Y-eq $script:lastY){$r.cursor_restored=([DeskFlow]::Move($cursor.X,$cursor.Y)-eq 1)}
 $null=[DeskFlow]::SetThreadDpiAwarenessContext($previousDpi)
 $r.product_sha256=(Get-FileHash $product).Hash.ToLowerInvariant();$r|ConvertTo-Json -Depth 6|Set-Content (Join-Path $out 'result.json') -Encoding utf8
}

} catch {
 $m=$_.Exception.Message.Replace($env:USERPROFILE,'<USERPROFILE>')
 @{result='failed';stage='outer-driver';error=$m;line=$_.InvocationInfo.ScriptLineNumber}|ConvertTo-Json|Set-Content (Join-Path $env:USERPROFILE 'asd-acceptance-20261122/instance-desktop-r2-20261126/result.json') -Encoding utf8
}
