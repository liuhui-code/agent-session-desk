$ErrorActionPreference='Stop'
if(Test-Path utf8-input.txt){[Console]::InputEncoding=[Text.UTF8Encoding]::new($false)}
Add-Type @"
using System;using System.Runtime.InteropServices;
public class VtProbeNative {
 [DllImport("kernel32.dll")] public static extern IntPtr GetStdHandle(int n);
 [DllImport("kernel32.dll")] public static extern bool GetConsoleMode(IntPtr h,out uint m);
 [DllImport("kernel32.dll")] public static extern uint GetConsoleCP();
}
"@
$p=Get-Process -Id $PID
@{pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content identity.json -Encoding utf8
'INPUT_READY'
$text=[Console]::ReadLine();$mode=[uint32]0;$ok=[VtProbeNative]::GetConsoleMode([VtProbeNative]::GetStdHandle(-10),[ref]$mode)
@{text=$text;units=@($text.ToCharArray()|ForEach-Object{[int]$_});input_mode=$mode;mode_available=$ok;input_codepage=[VtProbeNative]::GetConsoleCP()}|ConvertTo-Json|Set-Content normal-input.json -Encoding utf8
while(!(Test-Path seed-gate.txt)){Start-Sleep -Milliseconds 20}
Clear-Host
'SCROLL_NEEDLE_中文_α'
1..40|ForEach-Object{'history line '+$_}
'SEED_READY'
while(!(Test-Path alt-gate.txt)){Start-Sleep -Milliseconds 20}
[Console]::Write("`e[?1049h`e[2J`e[HVT_BOUNDARY_READY 中文 α")
$text=[Console]::ReadLine();$mode=[uint32]0;$ok=[VtProbeNative]::GetConsoleMode([VtProbeNative]::GetStdHandle(-10),[ref]$mode)
@{text=$text;units=@($text.ToCharArray()|ForEach-Object{[int]$_});columns=[Console]::WindowWidth;rows=[Console]::WindowHeight;input_mode=$mode;mode_available=$ok;input_codepage=[VtProbeNative]::GetConsoleCP()}|ConvertTo-Json|Set-Content alt-input.json -Encoding utf8
[Console]::Write("`e[?1049l")
'VT_BOUNDARY_DONE'
