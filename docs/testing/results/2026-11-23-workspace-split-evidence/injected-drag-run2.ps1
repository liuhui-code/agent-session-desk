$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'layout-20261123'
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName UIAutomationClient
Add-Type @'
using System;using System.Runtime.InteropServices;
public class AsdInjectDrag {
 [DllImport("user32.dll")] public static extern bool SetForegroundWindow(IntPtr h);
 [DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
 [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr h,out uint p);
 [DllImport("user32.dll")] public static extern IntPtr GetKeyboardLayout(uint thread);
 [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h,out RECT r);
 [DllImport("user32.dll")] public static extern bool MoveWindow(IntPtr h,int x,int y,int w,int height,bool redraw);
 [DllImport("user32.dll")] public static extern bool SetProcessDPIAware();
 [DllImport("user32.dll")] public static extern bool SetCursorPos(int x,int y);
 [DllImport("user32.dll")] public static extern short GetAsyncKeyState(int key);
 [DllImport("user32.dll",SetLastError=true)] public static extern uint SendInput(uint count,INPUT[] inputs,int size);
 [StructLayout(LayoutKind.Sequential)] public struct RECT {public int left,top,right,bottom;}
 [StructLayout(LayoutKind.Sequential)] public struct MOUSE {public int x,y;public uint data,flags,time;public UIntPtr extra;}
 [StructLayout(LayoutKind.Sequential)] public struct KEY {public ushort key,scan;public uint flags,time;public UIntPtr extra;}
 [StructLayout(LayoutKind.Explicit)] public struct DATA {[FieldOffset(0)]public MOUSE mouse;[FieldOffset(0)]public KEY key;}
 [StructLayout(LayoutKind.Sequential)] public struct INPUT {public uint type;public DATA data;}
 public static uint Text(string s){var a=new INPUT[s.Length*2];for(int i=0;i<s.Length;i++){a[i*2].type=1;a[i*2].data.key.scan=s[i];a[i*2].data.key.flags=4;a[i*2+1].type=1;a[i*2+1].data.key.scan=s[i];a[i*2+1].data.key.flags=6;}return SendInput((uint)a.Length,a,Marshal.SizeOf(typeof(INPUT)));}
 public static uint Key(ushort k){var a=new INPUT[2];a[0].type=1;a[0].data.key.key=k;a[1].type=1;a[1].data.key.key=k;a[1].data.key.flags=2;return SendInput(2,a,Marshal.SizeOf(typeof(INPUT)));}
 public static uint MouseButton(bool down){var a=new INPUT[1];a[0].data.mouse.flags=down?2u:4u;return SendInput(1,a,Marshal.SizeOf(typeof(INPUT)));}
 public static uint Move(int x,int y){var a=new INPUT[1];a[0].data.mouse.x=x;a[0].data.mouse.y=y;a[0].data.mouse.flags=0xC001;return SendInput(1,a,Marshal.SizeOf(typeof(INPUT)));}
 public static uint Click(){var a=new INPUT[2];a[0].data.mouse.flags=2;a[1].data.mouse.flags=4;return SendInput(2,a,Marshal.SizeOf(typeof(INPUT)));}
}
'@
[AsdInjectDrag]::SetProcessDPIAware()|Out-Null
function Guard {
 $ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json
 $p=Get-Process -Id $ready.pid
 if($p.Path -ne $ready.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Owned identity mismatch'}
 return $p
}
function Invoke-OwnedMouseMove($x,$y){
 $p=Guard;$foreground=[uint32]0;[AsdInjectDrag]::GetWindowThreadProcessId([AsdInjectDrag]::GetForegroundWindow(),[ref]$foreground)|Out-Null
 if($foreground -ne $p.Id){throw 'Foreground changed'}
 $screen=[Windows.Forms.SystemInformation]::VirtualScreen
 $nx=[int](($x-$screen.Left)*65535/($screen.Width-1));$ny=[int](($y-$screen.Top)*65535/($screen.Height-1))
 if([AsdInjectDrag]::Move($nx,$ny)-ne 1){throw 'Mouse move injection failed'}
}
try {
 $p=Guard;[AsdInjectDrag]::SetForegroundWindow($p.MainWindowHandle)|Out-Null;Start-Sleep -Milliseconds 300
 foreach($k in @(1,16,17,18,91,92)){if(([AsdInjectDrag]::GetAsyncKeyState($k)-band 32768)-ne 0){throw 'Button/modifier already held'}}
 $el=[Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle);$items=$el.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
 $rects=$null
 foreach($item in $items){$pat=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pat)){if($pat.DocumentRange.GetText(-1).StartsWith('Agent Session Desk')){$rects=$pat.DocumentRange.GetBoundingRectangles();break}}}
 if(!$rects -or $rects.Length -ne 32){throw 'Unexpected terminal geometry'}
 $line=$rects[10];$cell=$line.Width/120;$y=$line.Top+$line.Height/2
 Invoke-OwnedMouseMove ([int]($line.Left+49.5*$cell)) ([int]$y)
 if([AsdInjectDrag]::MouseButton($true)-ne 1){throw 'Mouse down failed'}
 try{for($step=1;$step -le 6;$step++){Invoke-OwnedMouseMove ([int]($line.Left+(49.5+11*$step/6)*$cell)) ([int]$y);Start-Sleep -Milliseconds 80}}finally{[AsdInjectDrag]::MouseButton($false)|Out-Null}
 @{result='injected';utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json|Set-Content (Join-Path $root 'injected-drag-run2.json') -Encoding utf8
}catch{$_|Out-String|Set-Content (Join-Path $root 'injected-drag-run2-error.txt') -Encoding utf8}
