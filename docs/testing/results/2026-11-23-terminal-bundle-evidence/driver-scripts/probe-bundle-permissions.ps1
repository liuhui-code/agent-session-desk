$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'bundle-desktop-20261123'
Add-Type @'
using System;using System.Runtime.InteropServices;
public class BundlePermissionProbe {
 [DllImport("kernel32.dll",SetLastError=true)] public static extern IntPtr OpenProcess(uint a,bool inherit,uint pid);
 [DllImport("kernel32.dll",SetLastError=true)] public static extern IntPtr CreateToolhelp32Snapshot(uint flags,uint pid);
 [DllImport("kernel32.dll")] public static extern bool CloseHandle(IntPtr h);
}
'@
$files=@(Get-ChildItem (Join-Path $root 'runtime/asd-bundle-desktop') -Filter '*.pid')
$rows=@();foreach($file in $files){$parts=([IO.File]::ReadAllText($file.FullName).Trim()).Split(':');$target=[uint32]$parts[0]
 $opened=@();foreach($access in @(0x1000,0x101000)){$h=[BundlePermissionProbe]::OpenProcess($access,$false,$target);$err=[Runtime.InteropServices.Marshal]::GetLastWin32Error();$opened+=,@{access=$access;success=($h -ne [IntPtr]::Zero);error=$err};if($h -ne [IntPtr]::Zero){[BundlePermissionProbe]::CloseHandle($h)|Out-Null}}
 $snapshot=[BundlePermissionProbe]::CreateToolhelp32Snapshot(0x18,$target);$errorCode=[Runtime.InteropServices.Marshal]::GetLastWin32Error();$ok=($snapshot.ToInt64() -ne -1);if($ok){[BundlePermissionProbe]::CloseHandle($snapshot)|Out-Null}
 $rows+=,@{server_pid=$target;server_creation=$parts[1];open_process=$opened;module_snapshot_ok=$ok;module_snapshot_error=$errorCode}
}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
@{utc=[DateTime]::UtcNow.ToString('o');is_administrator=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator);session=(Get-Process -Id $PID).SessionId;checks=$rows}|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'permission-probe-desktop.json') -Encoding utf8
