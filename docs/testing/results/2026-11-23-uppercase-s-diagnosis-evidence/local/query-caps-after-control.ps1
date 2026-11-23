$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'sidebar-key-diagnosis-20261123'
if(Test-Path (Join-Path $root 'caps-after-control.json')){Move-Item (Join-Path $root 'caps-after-control.json') (Join-Path $root 'caps-after-control-previous.json')}
$name='AsdSidebarKeyCapsAftercontrol20261123';if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Fresh query task required'}
$code=@'
$ErrorActionPreference='Stop'
Add-Type -AssemblyName System.Windows.Forms
Add-Type @"
using System;using System.Runtime.InteropServices;
public class KeyLayoutQuery{
 [DllImport("user32.dll")]public static extern int GetKeyboardLayoutList(int n,[Out] IntPtr[] values);
 [DllImport("user32.dll")]public static extern uint GetWindowThreadProcessId(IntPtr h,out uint p);
 [DllImport("user32.dll")]public static extern IntPtr GetKeyboardLayout(uint t);
 [DllImport("user32.dll")]public static extern IntPtr GetForegroundWindow();
 [DllImport("imm32.dll")]public static extern IntPtr ImmGetContext(IntPtr h);
 [DllImport("imm32.dll")]public static extern bool ImmReleaseContext(IntPtr h,IntPtr c);
 [DllImport("imm32.dll")]public static extern bool ImmGetOpenStatus(IntPtr c);
 [DllImport("imm32.dll")]public static extern bool ImmGetConversionStatus(IntPtr c,out uint conv,out uint sent);
}
"@
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'sidebar-key-diagnosis-20261123'
try{
 $ready=[IO.File]::ReadAllText((Join-Path $root 'desktop/ready.json'))|ConvertFrom-Json;$p=Get-Process -Id $ready.pid
 if($p.Path -ne $ready.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Original terminal changed'}
 $fore=[uint32]0;$null=[KeyLayoutQuery]::GetWindowThreadProcessId([KeyLayoutQuery]::GetForegroundWindow(),[ref]$fore);$target=[uint32]0;$thread=[KeyLayoutQuery]::GetWindowThreadProcessId($p.MainWindowHandle,[ref]$target);if($target -ne $p.Id){throw 'Owned window thread differs'}
 $n=[KeyLayoutQuery]::GetKeyboardLayoutList(0,$null);$layouts=[IntPtr[]]::new($n);$count=[KeyLayoutQuery]::GetKeyboardLayoutList($n,$layouts)
 $context=[KeyLayoutQuery]::ImmGetContext($p.MainWindowHandle);$open=$null;$conv=[uint32]0;$sent=[uint32]0;$conversion=$false
 if($context -ne [IntPtr]::Zero){try{$open=[KeyLayoutQuery]::ImmGetOpenStatus($context);$conversion=[KeyLayoutQuery]::ImmGetConversionStatus($context,[ref]$conv,[ref]$sent)}finally{[KeyLayoutQuery]::ImmReleaseContext($p.MainWindowHandle,$context)|Out-Null}}
 @{result='observed';utc=[DateTime]::UtcNow.ToString('o');thread=$thread;foreground_owned=($fore -eq $p.Id);capsLock=[Windows.Forms.Control]::IsKeyLocked([Windows.Forms.Keys]::CapsLock);pid=$p.Id;filetime=$ready.filetime;path=$p.Path;layout=[KeyLayoutQuery]::GetKeyboardLayout($thread).ToInt64();loaded_layouts=@($layouts|ForEach-Object{$_.ToInt64()});imm_context_available=($context -ne [IntPtr]::Zero);imm_open=$open;imm_conversion_available=$conversion;imm_conversion=$conv;imm_sentence=$sent}|ConvertTo-Json|Set-Content (Join-Path $root 'caps-after-control.json') -Encoding utf8
}catch{@{result='failed';error=$_.Exception.Message}|ConvertTo-Json|Set-Content (Join-Path $root 'caps-after-control.json') -Encoding utf8}
'@
$a=New-ScheduledTaskAction -Execute powershell.exe -Argument ('-NoProfile -NonInteractive -WindowStyle Hidden -STA -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code)))
$p=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask $name -Action $a -Principal $p -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 1))|Out-Null;Start-ScheduledTask $name
$file=Join-Path $root 'caps-after-control.json';$until=[DateTime]::UtcNow.AddSeconds(15);while(!(Test-Path $file) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if(!(Test-Path $file)){throw 'Query not observed'}
$until=[DateTime]::UtcNow.AddSeconds(5);while((Get-ScheduledTask $name).State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if((Get-ScheduledTask $name).State -eq 'Running'){throw 'Query still running'};Unregister-ScheduledTask $name -Confirm:$false
Get-Content $file -Raw -Encoding utf8
