param([string]$DiagnosticRoot,[string]$LegacyRoot,[string]$PackageRoot)
$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$d=$DiagnosticRoot
$p=$PackageRoot
$legacy=$LegacyRoot
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $legacy 'desktop-api.ps1'))))
Add-Type @'
using System;using System.Runtime.InteropServices;
public class IdleNative{
 [StructLayout(LayoutKind.Sequential)]public struct LastInfo{public uint Size,Time;}
 [DllImport("user32.dll")]public static extern bool GetLastInputInfo(ref LastInfo info);
 public static uint LastTick(){var x=new LastInfo();x.Size=(uint)Marshal.SizeOf(typeof(LastInfo));if(!GetLastInputInfo(ref x))throw new Exception("Last input unavailable");return x.Time;}

 [DllImport("user32.dll")]public static extern bool ShowWindowAsync(IntPtr h,int c);
 [DllImport("user32.dll")]public static extern bool IsIconic(IntPtr h);
 [DllImport("user32.dll")]public static extern bool IsWindow(IntPtr h);
 [DllImport("user32.dll")]public static extern IntPtr GetForegroundWindow();
 [DllImport("user32.dll")]public static extern bool SetForegroundWindow(IntPtr h);
 [DllImport("user32.dll")]public static extern bool PostMessage(IntPtr h,uint m,IntPtr w,IntPtr l);
 [DllImport("user32.dll")]public static extern uint GetWindowThreadProcessId(IntPtr h,out uint p);
}
'@
$log=Join-Path $d 'state.jsonl';$stages=[Collections.Generic.List[object]]::new();$created=@();$fixtureHandles=@();$manager=$null;$window=[IntPtr]::Zero;$hostProcess=$null
$r=@{result='started';forced=0;clipboard_apis=0;keyboard_apis=0;baseline_ready=$false}
$original=[IdleNative]::GetForegroundWindow();[uint32]$originalPid=0;[IdleNative]::GetWindowThreadProcessId($original,[ref]$originalPid)|Out-Null
$originalProcess=$null;try{$originalProcess=Get-Process -Id $originalPid;$null=$originalProcess.Handle}catch{$originalProcess=$null}
$lastInputTick=[IdleNative]::LastTick()
function Read-Json([string]$path,[switch]$First){
 try{
  $f=[IO.File]::Open($path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite-bor [IO.FileShare]::Delete));$reader=[IO.StreamReader]::new($f)
  try{$text=$reader.ReadToEnd()}finally{$reader.Dispose()}
  if($First){$text=($text-split "`n")[0]}
  if(!$First-and $path.EndsWith('state.jsonl')){foreach($line in @($text-split "`n"|Where-Object{$_})[-1..0]){try{return $line|ConvertFrom-Json}catch{}};return $null}
  return $text|ConvertFrom-Json
 }catch{return $null}
}
$petWindow=[IntPtr]::Zero;$lastPet=''
function Observe-Sleep([int]$seconds){
 $until=[DateTime]::UtcNow.AddSeconds($seconds)
 do{
  if($petWindow-ne [IntPtr]::Zero){$caption=[DeskFlow]::Title($petWindow);if($caption-cne $script:lastPet){$stages.Add(@{stage='pet';at=[DateTime]::UtcNow.ToString('o');caption=$caption});$script:lastPet=$caption}}
  Start-Sleep -Milliseconds 250
 }while([DateTime]::UtcNow-lt $until)
}
function Record([string]$stage){$stages.Add(@{stage=$stage;at=[DateTime]::UtcNow.ToString('o');iconic=($window-ne [IntPtr]::Zero-and [IdleNative]::IsIconic($window))})}
function Wait-State([scriptblock]$test,[int]$seconds,[string]$name){$until=[DateTime]::UtcNow.AddSeconds($seconds);do{if(& $test){return};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow-lt $until);throw ('Deadline: '+$name)}
function Tabs-Safe{
 $element=[Windows.Automation.AutomationElement]::FromHandle($window)
 $tabs=$element.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem))
 if($tabs.Count-gt 1){throw 'Owned window gained other tabs; no window close'}
}
try{
 $m=Get-Content -Raw (Join-Path $d 'hashes.json')|ConvertFrom-Json
 foreach($name in @('asd-idle-diag.exe','asd-idle-fixture.exe')){if((Get-FileHash (Join-Path $d $name)).Hash.ToLowerInvariant()-cne $m.$name){throw 'Diagnostic hash changed'}}
 $hostInfo=Get-Content -Raw (Join-Path $p 'local-host.json')|ConvertFrom-Json
 if((Get-FileHash $hostInfo.terminal.path).Hash.ToLowerInvariant()-cne $hostInfo.terminal.sha256){throw 'Pinned host changed'}
 Copy-Item (Join-Path $p 'local-host.json') (Join-Path $d 'local-host.json')
 $bundle=Join-Path $p 'terminal-runtime'
 if((Get-FileHash (Join-Path $bundle 'psmux.exe')).Hash.ToLowerInvariant()-cne 'ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b'){throw 'Pinned backend changed'}
 $exe=Join-Path $d 'asd-idle-diag.exe';$fixture=Join-Path $d 'asd-idle-fixture.exe'
 $g=@('--terminal-bundle',$bundle,'--runtime-dir',(Join-Path $d 'runtime'),'--namespace','asd-idlerestore-r3','--state-dir',(Join-Path $d 'state'))
 foreach($name in @('IdleA','IdleB')){
  $journal=Join-Path $d ($name+'.jsonl')
  $value=& $exe @g start --name $name --cwd $d --env ('ASD_FIXTURE_REPORT_PATH='+$journal) -- $fixture 2>&1
  if($LASTEXITCODE-ne 0){throw ('Fixture start failed: '+($value|Out-String))}
  $created+=($value|Out-String|ConvertFrom-Json)
  Wait-State {$null-ne (Read-Json $journal -First)} 10 'fixture report'
  $receipt=Read-Json $journal -First
  $process=Get-Process -Id $receipt.pid;$null=$process.Handle
  if($process.Path-cne $fixture){throw 'Fixture image mismatch'}
  $fixtureHandles+=$process
 }
 $created|ConvertTo-Json -Depth 6|Set-Content (Join-Path $d 'created.json')
 $env:ASD_IDLE_DIAGNOSTIC=$log
 Record 'launch'
 & $exe @g --pet *> (Join-Path $d 'launcher.log');$r.launch_code=$LASTEXITCODE
 if($LASTEXITCODE-ne 0){throw 'Workbench launcher failed'}
 Wait-State {$null-ne (Read-Json ($log+'.window.json'))} 15 'bound window receipt'
 $bound=Read-Json ($log+'.window.json')
 $manager=Get-Process -Id $bound.manager_pid;$null=$manager.Handle
 if($manager.Path-cne $exe){throw 'Manager image mismatch'}
 $window=[IntPtr][long]$bound.hwnd;[uint32]$hostPid=0;[IdleNative]::GetWindowThreadProcessId($window,[ref]$hostPid)|Out-Null
 $hostProcess=Get-Process -Id $hostPid;$null=$hostProcess.Handle
 if(![IdleNative]::IsWindow($window)-or $hostProcess.Path-cne $hostInfo.terminal.path){throw 'Bound host mismatch'}
 Tabs-Safe
 Wait-State {if(!(Test-Path $log)){return $false};$tail=Read-Json $log;return $tail.conn -and $tail.sessions-eq 2} 30 'initial native terminal'
 $r.baseline_ready=$true;$pets=@([DeskFlow]::Find('ASD Blocking Reminders',$manager.Id));if($pets.Count-eq 1){$petWindow=$pets[0];$r.pet_observed=$true};Record 'baseline';Observe-Sleep 10
 foreach($cycle in 1..2){
  if($manager.HasExited){throw 'Manager exited before minimize'}
  Record ('minimize-'+$cycle)
  if(![IdleNative]::ShowWindowAsync($window,6)){throw 'Minimize request rejected'}
  Wait-State {[IdleNative]::IsIconic($window)} 3 'actual minimize'
  Record ('minimized-'+$cycle);Observe-Sleep 4
  Record ('restore-'+$cycle)
  if(![IdleNative]::ShowWindowAsync($window,9)){throw 'Restore request rejected'}
  Wait-State {![IdleNative]::IsIconic($window)} 3 'actual restore'
  Record ('restored-'+$cycle);Observe-Sleep 12
 }
 Record 'completed';$r.result='completed';$r.manager_pid=$manager.Id;$r.manager_start=$manager.StartTime.ToUniversalTime().ToString('o')
}catch{$r.result='failed';$r.error=$_.Exception.Message}
finally{
 try{
  if($manager-and !$manager.HasExited-and $window-ne [IntPtr]::Zero){
   Tabs-Safe;[uint32]$nowPid=0;[IdleNative]::GetWindowThreadProcessId($window,[ref]$nowPid)|Out-Null
   if($nowPid-ne $hostProcess.Id-or $hostProcess.HasExited){throw 'Window identity changed'}
   [IdleNative]::PostMessage($window,0x10,[IntPtr]::Zero,[IntPtr]::Zero)|Out-Null
   if(!$manager.WaitForExit(15000)){throw 'Owned UI did not exit; no force stop'}
   $r.manager_exit=$manager.ExitCode
  }
 }catch{$r.close_error=$_.Exception.Message}
 $stops=@();foreach($item in $created){$null=& $exe @g stop $item.session_id 2>&1;$stops+=@{session=$item.session_id;code=$LASTEXITCODE}}
 $r.stops=$stops;$ends=@();foreach($process in $fixtureHandles){$ends+=@{pid=$process.Id;start=$process.StartTime.ToUniversalTime().ToString('o');ended=$process.WaitForExit(10000)}};$r.fixture_ends=$ends
 if($originalProcess-and !$originalProcess.HasExited-and [IdleNative]::IsWindow($original)){
  $nowTick=[IdleNative]::LastTick()
  if($nowTick-eq $lastInputTick){[IdleNative]::SetForegroundWindow($original)|Out-Null;$r.foreground_restored=([IdleNative]::GetForegroundWindow()-eq $original)}else{$r.foreground_restore='skipped: input changed'}
 }
 $stages|ConvertTo-Json -Depth 5|Set-Content (Join-Path $d 'stages.json')
 $r|ConvertTo-Json -Depth 7|Set-Content (Join-Path $d 'result.json')
}
