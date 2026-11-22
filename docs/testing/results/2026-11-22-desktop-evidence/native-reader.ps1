$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';Set-Location $b
$source=@'
using System;using System.IO;using System.Text;using System.Runtime.InteropServices;using System.Diagnostics;using System.Web.Script.Serialization;
class NativeReader {
 [DllImport("kernel32.dll")] static extern IntPtr GetStdHandle(int n);
 [DllImport("kernel32.dll",CharSet=CharSet.Unicode,SetLastError=true)] static extern bool ReadConsoleW(IntPtr h,[Out] char[] b,uint n,out uint r,IntPtr reserved);
 [DllImport("kernel32.dll",CharSet=CharSet.Unicode)] static extern bool WriteConsoleW(IntPtr h,string s,uint n,out uint r,IntPtr reserved);
 static void Log(object o){File.AppendAllText(Environment.GetEnvironmentVariable("ASD_TEST_JOURNAL"),new JavaScriptSerializer().Serialize(o)+"\n",new UTF8Encoding(false));}
 static void Print(string s){uint n;WriteConsoleW(GetStdHandle(-11),s,(uint)s.Length,out n,IntPtr.Zero);}
 static int Main(){Log(new{kind="started",pid=Process.GetCurrentProcess().Id,api="ReadConsoleW"});Print("ASD native Unicode fixture ready\r\n");while(true){var b=new char[4096];uint n;if(!ReadConsoleW(GetStdHandle(-10),b,4096,out n,IntPtr.Zero))return Marshal.GetLastWin32Error();if(n==0)return 0;string s=new string(b,0,(int)n).TrimEnd('\r','\n');Log(new{kind="input",line=s});if(s=="exit")return 0;Print("ASD native received: "+s+"\r\n");}}
}
'@
Add-Type -TypeDefinition $source -ReferencedAssemblies 'System.Web.Extensions.dll' -OutputAssembly "$b\desktop\native-reader.exe" -OutputType ConsoleApplication
$common=@('--backend',"$b\backend\psmux.exe",'--runtime-dir',"$b\desktop-runtime",'--namespace','asd-desktop-20261122')
$asd="$b\product\asd-windows-amd64.exe"
function Flat($x){if($x -is [Array]){foreach($v in $x){Flat $v}}elseif($x.PSObject.Properties['session_id']){$x}elseif($x.PSObject.Properties['value']){Flat $x.value}else{throw 'Unexpected identity record'}}
$old=@(Flat (Get-Content desktop/created.json -Raw | ConvertFrom-Json))
$live=& $asd @common list --json | ConvertFrom-Json
foreach($c in $old){$x=@($live|Where-Object session_id -eq $c.session_id);if($x.Count-ne 1 -or $x[0].run_id-ne$c.run_id){throw 'Run changed after manager q'}}
$live|ConvertTo-Json -Depth 8|Set-Content desktop/after-q-list.json -Encoding UTF8
$s=& $asd @common start --name 'Unicode 原生 E' --env ('ASD_TEST_JOURNAL='+"$b\desktop\fixture-E.jsonl") --cwd $b -- "$b\desktop\native-reader.exe"
if($LASTEXITCODE-ne 0){throw 'Native reader create failed'}
$old+=,($s|ConvertFrom-Json);$old|ConvertTo-Json -Depth 8|Set-Content desktop/created.json -Encoding UTF8
@{source='synthetic ReadConsoleW reader; not cmd/asd-fixture';sha256=(Get-FileHash desktop/native-reader.exe).Hash.ToLowerInvariant();session=($s|ConvertFrom-Json)}|ConvertTo-Json -Depth 6
