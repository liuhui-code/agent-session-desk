$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$name='AsdMainAllMotionMove3_20261123'
if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Move probe already exists'}
$code=@'
$step='initialization'
try{
$ErrorActionPreference='Stop';$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-allmotion-desktop-20261123'
Add-Type -AssemblyName System.Windows.Forms
Add-Type @"
using System;using System.Runtime.InteropServices;
public class MouseMoveProbe {
[DllImport("user32.dll")]public static extern bool SetProcessDPIAware();
[DllImport("user32.dll")]public static extern bool SetForegroundWindow(IntPtr h);
[DllImport("user32.dll")]public static extern IntPtr GetForegroundWindow();
[DllImport("user32.dll")]public static extern uint GetWindowThreadProcessId(IntPtr h,out uint p);
[DllImport("user32.dll")]public static extern uint SendInput(uint n,INPUT[] input,int size);
[StructLayout(LayoutKind.Sequential)]public struct MOUSE {public int x,y;public uint data,flags,time;public UIntPtr extra;}
[StructLayout(LayoutKind.Sequential)]public struct INPUT {public uint type;public MOUSE mouse;}
public static uint Move(int x,int y){var a=new INPUT[1];a[0].mouse.x=x;a[0].mouse.y=y;a[0].mouse.flags=0xC001;return SendInput(1,a,Marshal.SizeOf(typeof(INPUT)));}
}
"@
[MouseMoveProbe]::SetProcessDPIAware()|Out-Null
$step='window/foreground guard'
$ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json;$p=Get-Process -Id $ready.pid
if($p.Path -ne $ready.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Original window changed'}
[MouseMoveProbe]::SetForegroundWindow($p.MainWindowHandle)|Out-Null;Start-Sleep -Milliseconds 200
$fg=[uint32]0;$null=[MouseMoveProbe]::GetWindowThreadProcessId([MouseMoveProbe]::GetForegroundWindow(),[ref]$fg);if($fg -ne $p.Id){throw 'Foreground not owned; no input'}
$step='row metadata'
$all=@(Get-ChildItem (Join-Path $root 'desktop-continuation/responses') -Filter '*.json'|Sort-Object Name|ForEach-Object{$r=[IO.File]::ReadAllText($_.FullName)|ConvertFrom-Json;if($r.command.kind -eq 'divider'){$r}});$r=$all[-1];$rect=$r.rowRectangle;$screen=[Windows.Forms.SystemInformation]::VirtualScreen
$step='native SendInput'
$points=@();foreach($col in @(43,0)){$x=[int]($rect[0]+($col+0.5)*$rect[2]/120);$y=[int]($rect[1]+$rect[3]/2);$nx=[int](($x-$screen.Left)*65535/($screen.Width-1));$ny=[int](($y-$screen.Top)*65535/($screen.Height-1));$sent=[MouseMoveProbe]::Move($nx,$ny);if($sent -ne 1){throw 'Native move rejected'};$points+=,@{column=$col;point=@($x,$y);sent=$sent};Start-Sleep -Milliseconds 400}
@{utc=[DateTime]::UtcNow.ToString('o');result='observed';points=$points;trace=[IO.File]::ReadAllText((Join-Path $root 'mouse-trace.log'));scope='Mouse-only absolute SendInput over the registered synthetic divider, no keyboard/clipboard/content capture'}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'native-move-probe3.json') -Encoding utf8
}catch{@{result='failed';step=$step;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $env:USERPROFILE 'asd-acceptance-20261122/main-allmotion-desktop-20261123/native-move-probe3.json') -Encoding utf8}
'@
$a=New-ScheduledTaskAction -Execute powershell.exe -Argument ('-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code)));$p=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask $name -Action $a -Principal $p -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 1))|Out-Null;Start-ScheduledTask $name
$f=Join-Path $base 'main-allmotion-desktop-20261123/native-move-probe3.json';$until=[DateTime]::UtcNow.AddSeconds(15);while(!(Test-Path $f) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $f)){throw 'Move probe not observed'}
Get-Content $f -Raw -Encoding utf8
$until=[DateTime]::UtcNow.AddSeconds(5);while((Get-ScheduledTask $name).State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if((Get-ScheduledTask $name).State -eq 'Running'){throw 'Move probe still running'};Unregister-ScheduledTask $name -Confirm:$false
