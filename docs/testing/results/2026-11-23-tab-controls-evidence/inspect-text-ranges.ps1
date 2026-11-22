$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tab-controls-20261123'
Add-Type -AssemblyName UIAutomationClient
Add-Type @'
using System;using System.Runtime.InteropServices;
public class AsdRangeInspect {[DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr h);}
'@
try {
 $ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json
 $p=Get-Process -Id $ready.pid
 if($p.Path -ne $ready.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Owned identity mismatch'}
 $el=[Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
 $items=$el.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
 $results=@()
 foreach($item in $items){$pattern=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){
  $range=$pattern.DocumentRange;$rect=$item.Current.BoundingRectangle
  $v=@{name=$item.Current.Name;text=$range.GetText(-1);controlRect=@($rect.Left,$rect.Top,$rect.Width,$rect.Height)}
  try{$v.rangeRects=@($range.GetBoundingRectangles()|ForEach-Object{@($_.Left,$_.Top,$_.Width,$_.Height)})}catch{$v.rangeError=$_.Exception.Message}
  $results+=,$v
 }}
 $results|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'text-ranges.json') -Encoding utf8
 [AsdRangeInspect]::SetForegroundWindow($p.MainWindowHandle)|Out-Null
}catch{$_|Out-String|Set-Content (Join-Path $root 'text-ranges-error.txt') -Encoding utf8}
