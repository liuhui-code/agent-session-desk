$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'visual-compare-20261123'
$frames=Join-Path $root 'frames';$manifest=[IO.File]::ReadAllText((Join-Path $frames 'manifest.json'))|ConvertFrom-Json
Add-Type @'
using System;using System.Runtime.InteropServices;
public class AsdReplayConsole {
 [DllImport("kernel32.dll",SetLastError=true)] public static extern IntPtr GetStdHandle(int n);
 [DllImport("kernel32.dll",SetLastError=true)] public static extern bool GetConsoleMode(IntPtr h,out uint mode);
 [DllImport("kernel32.dll",SetLastError=true)] public static extern bool SetConsoleMode(IntPtr h,uint mode);
}
'@
$outHandle=[AsdReplayConsole]::GetStdHandle(-11);$inHandle=[AsdReplayConsole]::GetStdHandle(-10)
$outMode=[uint32]0;$inMode=[uint32]0
if(![AsdReplayConsole]::GetConsoleMode($outHandle,[ref]$outMode)-or ![AsdReplayConsole]::GetConsoleMode($inHandle,[ref]$inMode)){throw 'Replay requires a real console'}
if(![AsdReplayConsole]::SetConsoleMode($outHandle,($outMode -bor 12))){throw 'VT output mode failed'}
if(![AsdReplayConsole]::SetConsoleMode($inHandle,($inMode -band (-bnot 6)))){[AsdReplayConsole]::SetConsoleMode($outHandle,$outMode)|Out-Null;throw 'Non-echo input mode failed'}
$stdout=[Console]::OpenStandardOutput();$esc=[char]27
function Raw($value){$b=[Text.Encoding]::UTF8.GetBytes($value);$stdout.Write($b,0,$b.Length);$stdout.Flush()}
function Receipt($value){
 $temp=Join-Path $root 'frame-receipt.tmp';$target=Join-Path $root 'frame-receipt.json'
 [IO.File]::WriteAllText($temp,($value|ConvertTo-Json -Depth 4),[Text.UTF8Encoding]::new($false));[IO.File]::Move($temp,$target,$true)
}
try{
 Raw ($esc+'[?1049h'+$esc+'[?25l')
 $p=[Diagnostics.Process]::GetCurrentProcess()
 @{pid=$p.Id;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.MainModule.FileName;columns=[Console]::WindowWidth;rows=[Console]::WindowHeight}|ConvertTo-Json|Set-Content (Join-Path $root 'replay-ready.json') -Encoding utf8
 $last='';$until=[DateTime]::UtcNow.AddMinutes(20)
 while([DateTime]::UtcNow -lt $until -and !(Test-Path (Join-Path $root 'replay.stop'))){
  $request=Join-Path $root 'frame-request.txt'
  if(!(Test-Path $request)){Start-Sleep -Milliseconds 50;continue}
  $name=[IO.File]::ReadAllText($request).Trim();$width=[Console]::WindowWidth;$height=[Console]::WindowHeight;$key="$name/$width/$height"
  if($key -eq $last){Start-Sleep -Milliseconds 50;continue};$last=$key
  $property=$manifest.PSObject.Properties[$name]
  if(!$property){Receipt @{ok=$false;frame=$name;error='Unknown frame'};continue}
  $frame=$property.Value
  if($width -ne $frame.width -or $height -ne $frame.height){Receipt @{ok=$false;frame=$name;columns=$width;rows=$height;error='Console dimensions do not match original capture'};continue}
  if([IO.Path]::GetFileName($frame.file) -ne $frame.file){throw 'Invalid manifest path'}
  $bytes=[IO.File]::ReadAllBytes((Join-Path $frames $frame.file));if($bytes.Length -gt 2097152){throw 'Frame exceeds defensive replay bound'}
  $stdout.Write($bytes,0,$bytes.Length);$stdout.Flush()
  Receipt @{ok=$true;frame=$name;columns=$width;rows=$height;sha256=(Get-FileHash (Join-Path $frames $frame.file)).Hash;source='original upstream tmux cell capture; serialized VT replay'}
 }
}finally{
 Raw ($esc+'[0m'+$esc+'[?25h'+$esc+'[?1049l')
 [AsdReplayConsole]::SetConsoleMode($inHandle,$inMode)|Out-Null;[AsdReplayConsole]::SetConsoleMode($outHandle,$outMode)|Out-Null
 @{utc=[DateTime]::UtcNow.ToString('o');stopped=(Test-Path (Join-Path $root 'replay.stop'))}|ConvertTo-Json|Set-Content (Join-Path $root 'replay-finished.json') -Encoding utf8
}
