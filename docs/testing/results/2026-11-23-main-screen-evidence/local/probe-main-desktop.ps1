$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$name='AsdMainParityRead20261123'
if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Read-only probe already exists'}
$code=@'
$ErrorActionPreference='Stop';$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
Add-Type @"
using System;using System.Text;using System.Runtime.InteropServices;
public class DesktopRead {
[DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();
[DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr h,out uint p);
[DllImport("user32.dll",SetLastError=true)] public static extern IntPtr OpenInputDesktop(uint flags,bool inherit,uint access);
[DllImport("user32.dll",SetLastError=true,CharSet=CharSet.Unicode)] public static extern bool GetUserObjectInformation(IntPtr h,int index,StringBuilder data,uint length,out uint needed);
[DllImport("user32.dll")] public static extern bool CloseDesktop(IntPtr h);
}
"@
$own=(Get-Process -Id $PID).SessionId;$h=[DesktopRead]::OpenInputDesktop(0,$false,1);$desktop='unavailable';$err=[Runtime.InteropServices.Marshal]::GetLastWin32Error()
if($h -ne [IntPtr]::Zero){try{$s=[Text.StringBuilder]::new(256);$needed=[uint32]0;if([DesktopRead]::GetUserObjectInformation($h,2,$s,512,[ref]$needed)){$desktop=$s.ToString();$err=0}}finally{[DesktopRead]::CloseDesktop($h)|Out-Null}}
$fg=[uint32]0;$null=[DesktopRead]::GetWindowThreadProcessId([DesktopRead]::GetForegroundWindow(),[ref]$fg);$category='none';if($fg){$category='other';$p=Get-Process -Id $fg -ErrorAction SilentlyContinue;if($p -and $p.Path -eq (Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe')){$category='owned-terminal'}}
@{utc=[DateTime]::UtcNow.ToString('o');session=$own;input_desktop=$desktop;desktop_error=$err;foreground_category=$category;logonui_count=@(Get-Process LogonUI -ErrorAction SilentlyContinue|Where-Object{$_.SessionId -eq $own}).Count;scope='Read-only desktop/session availability; no focus, keyboard, mouse, clipboard or screen capture'}|ConvertTo-Json|Set-Content (Join-Path $base 'main-parity-desktop-probe.json') -Encoding utf8
'@
$a=New-ScheduledTaskAction -Execute powershell.exe -Argument ('-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code)))
$p=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask $name -Action $a -Principal $p -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 1))|Out-Null;Start-ScheduledTask $name
$f=Join-Path $base 'main-parity-desktop-probe.json';$until=[DateTime]::UtcNow.AddSeconds(15);while(!(Test-Path $f) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $f)){throw 'Probe not observed'}
Get-Content $f -Raw -Encoding utf8
$until=[DateTime]::UtcNow.AddSeconds(5);while((Get-ScheduledTask $name).State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if((Get-ScheduledTask $name).State -eq 'Running'){throw 'Probe still running'};Unregister-ScheduledTask $name -Confirm:$false
