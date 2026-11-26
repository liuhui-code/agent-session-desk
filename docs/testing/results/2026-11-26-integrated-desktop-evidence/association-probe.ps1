$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.Encoding]::UTF8
Add-Type @'
using System;using System.Runtime.InteropServices;
public class AssocProbe {
 [DllImport("shlwapi.dll",CharSet=CharSet.Unicode)]public static extern uint AssocQueryString(uint flags,uint kind,string type,string extra,IntPtr output,ref uint length);
 public static object Probe(string ext,uint kind,uint flags){uint n=0;uint hr=AssocQueryString(flags,kind,ext,"open",IntPtr.Zero,ref n);return new {result=hr.ToString("x8"),required_chars=n};}
}
'@
@{md_prog=@([AssocProbe]::Probe('.md',20,0),[AssocProbe]::Probe('.md',20,0x400));md_exec=@([AssocProbe]::Probe('.md',2,0),[AssocProbe]::Probe('.md',2,0x400));unknown_prog=[AssocProbe]::Probe('.asd-unregistered-20261126',20,0x400);registry_writes=0;applications_launched=0}|ConvertTo-Json -Depth 4
