$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tabs-ui-20261123/desktop'
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName UIAutomationClient
Add-Type @'
using System;using System.Runtime.InteropServices;
public class AsdPasteProbe { [DllImport("user32.dll")]public static extern bool SetForegroundWindow(IntPtr h); [DllImport("user32.dll")]public static extern IntPtr GetForegroundWindow();[DllImport("user32.dll")]public static extern uint GetWindowThreadProcessId(IntPtr h,out uint p);}
'@
function Guard {
 $ready=[IO.File]::ReadAllText((Join-Path $root 'ready.json'))|ConvertFrom-Json
 $p=Get-Process -Id $ready.pid
 $f=[uint32]0;[AsdPasteProbe]::GetWindowThreadProcessId([AsdPasteProbe]::GetForegroundWindow(),[ref]$f)|Out-Null
 if($p.Id -ne $f -or $p.Path -ne $ready.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Owned foreground identity mismatch'}
 return $p
}
$r=[IO.File]::ReadAllText((Join-Path $root 'ready.json'))|ConvertFrom-Json
$p=Get-Process -Id $r.pid
if($p.Path -ne $r.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $r.filetime){throw 'Target identity mismatch'}
[AsdPasteProbe]::SetForegroundWindow($p.MainWindowHandle)|Out-Null
Start-Sleep -Milliseconds 300
$saved=[Windows.Forms.Clipboard]::GetDataObject();$results=@()
try {
 foreach($keys in @('^v','^+v','+{INSERT}')) {
  $p=Guard
  [Windows.Forms.Clipboard]::SetText("Write-Output 'paste probe'")
  [Windows.Forms.SendKeys]::SendWait($keys)
  Start-Sleep -Milliseconds 1200
  $p=Guard
  $el=[Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
  $items=$el.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
  $text=@();foreach($item in $items){$pat=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pat)){$text+=,$pat.DocumentRange.GetText(-1)}}
  $results+=,@{keys=$keys;text=($text -join "`n");clipboard=[Windows.Forms.Clipboard]::GetText()}
  [Windows.Forms.SendKeys]::SendWait('^c');Start-Sleep -Milliseconds 300
 }
}finally{if($saved){[Windows.Forms.Clipboard]::SetDataObject($saved,$true)}else{[Windows.Forms.Clipboard]::Clear()};$results|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'paste-probe.json') -Encoding utf8}
