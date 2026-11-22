try {
Add-Type -AssemblyName System.Windows.Forms
$ErrorActionPreference='Stop'
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/desktop'
Add-Type -AssemblyName System.Windows.Forms
Add-Type @'
using System;using System.Runtime.InteropServices;public class AsdPasteGuard{[DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();[DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr h,out uint p);}
'@
$p=[uint32]0;[AsdPasteGuard]::GetWindowThreadProcessId([AsdPasteGuard]::GetForegroundWindow(),[ref]$p)|Out-Null
$ready=Get-Content (Join-Path $b 'ready.json') -Raw | ConvertFrom-Json
if($p -ne $ready.pid){throw 'Paste foreground guard failed'}
$old=[Windows.Forms.Clipboard]::GetDataObject()
try {
 [Windows.Forms.Clipboard]::SetText("ASD_PASTE_1 中文`r`nASD_PASTE_2 🧪")
 [Windows.Forms.SendKeys]::SendWait('^v')
 Start-Sleep -Milliseconds 500
 @{ok=$true;pid=$p;input='ASD_PASTE_1 中文\r\nASD_PASTE_2 🧪';clipboard_restored=$true} | ConvertTo-Json | Set-Content (Join-Path $b 'paste-result.json') -Encoding UTF8
}finally{if($old){[Windows.Forms.Clipboard]::SetDataObject($old,$true)}else{[Windows.Forms.Clipboard]::Clear()}}

} catch { @{error=$_.Exception.Message;stack=$_.ScriptStackTrace} | ConvertTo-Json | Set-Content (Join-Path $env:USERPROFILE 'asd-acceptance-20261122/desktop/paste-error.json') -Encoding UTF8 }
