$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tabs-ui-20261123/desktop'
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName UIAutomationClient
Add-Type @'
using System;using System.Runtime.InteropServices;
public class AsdScanProbe {
 [DllImport("user32.dll")] public static extern short GetKeyState(int key);
 [DllImport("user32.dll")] public static extern uint MapVirtualKey(uint code,uint type);
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
 public static uint Key(ushort k){var a=new INPUT[2];a[0].type=1;a[0].data.key.scan=(ushort)MapVirtualKey(k,0);a[0].data.key.flags=8;a[1].type=1;a[1].data.key.scan=(ushort)MapVirtualKey(k,0);a[1].data.key.flags=10;return SendInput(2,a,Marshal.SizeOf(typeof(INPUT)));}
 public static uint Click(){var a=new INPUT[2];a[0].data.mouse.flags=2;a[1].data.mouse.flags=4;return SendInput(2,a,Marshal.SizeOf(typeof(INPUT)));}
}
'@
function Guard {
 $ready=[IO.File]::ReadAllText((Join-Path $root 'ready.json'))|ConvertFrom-Json
 $p=Get-Process -Id $ready.pid
 $f=[uint32]0;[AsdScanProbe]::GetWindowThreadProcessId([AsdScanProbe]::GetForegroundWindow(),[ref]$f)|Out-Null
 if($p.Id -ne $f -or $p.Path -ne $ready.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Owned foreground identity mismatch'}
 return $p
}
$r=[IO.File]::ReadAllText((Join-Path $root 'ready.json'))|ConvertFrom-Json
$p=Get-Process -Id $r.pid
if($p.Path -ne $r.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $r.filetime){throw 'Target identity mismatch'}
[AsdScanProbe]::SetForegroundWindow($p.MainWindowHandle)|Out-Null
Start-Sleep -Milliseconds 300
$before=[AsdScanProbe]::GetKeyState(20)
if(($before-band 1)-ne 0){Guard|Out-Null;[AsdScanProbe]::Key(20)|Out-Null}
foreach($ch in 'NIHAO'.ToCharArray()){Guard|Out-Null;[AsdScanProbe]::Key([uint16][char]$ch)|Out-Null;Start-Sleep -Milliseconds 150}
@{caps_before=$before;caps_after=[AsdScanProbe]::GetKeyState(20);utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json|Set-Content (Join-Path $root 'scan-probe.json') -Encoding utf8
