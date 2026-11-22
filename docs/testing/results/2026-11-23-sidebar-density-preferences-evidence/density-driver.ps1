$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$out=Join-Path $base 'density-20261123/desktop'
New-Item -ItemType Directory -Path $out -ErrorAction Stop|Out-Null
foreach($d in @('commands','responses','done')){New-Item -ItemType Directory -Path (Join-Path $out $d)|Out-Null}
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
Add-Type -AssemblyName UIAutomationClient
Add-Type @'
using System;using System.Runtime.InteropServices;
public class AsdDesktop {
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
 public static uint Click(){var a=new INPUT[2];a[0].data.mouse.flags=2;a[1].data.mouse.flags=4;return SendInput(2,a,Marshal.SizeOf(typeof(INPUT)));}
}
'@
[AsdDesktop]::SetProcessDPIAware() | Out-Null
$script:ownedId=0
function Guard {
 $p=Get-Process -Id $script:ownedId
 if($p.MainWindowHandle -eq 0 -or $p.Path -ne $script:terminal -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $script:ownedFiletime){throw 'Owned terminal identity mismatch'}
 $foreground=[uint32]0;[AsdDesktop]::GetWindowThreadProcessId([AsdDesktop]::GetForegroundWindow(),[ref]$foreground)|Out-Null
 if($foreground -ne $p.Id){throw 'Foreground changed; input refused'}
 return $p
}
function State {
 $p=Guard;$r=New-Object AsdDesktop+RECT;[AsdDesktop]::GetWindowRect($p.MainWindowHandle,[ref]$r)|Out-Null
 $root=[Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
 $items=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
 $texts=@();$tree=@()
 foreach($el in $items){$name=[string]$el.Current.Name;$tree+=@{name=$name;type=$el.Current.ControlType.ProgrammaticName};$pat=$null;if($el.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pat)){$texts+=,$pat.DocumentRange.GetText(-1)}}
 $f=[uint32]0;$thread=[AsdDesktop]::GetWindowThreadProcessId([AsdDesktop]::GetForegroundWindow(),[ref]$f)
 return @{pid=$p.Id;filetime=$script:ownedFiletime;path=$p.Path;session=$p.SessionId;rect=@($r.left,$r.top,($r.right-$r.left),($r.bottom-$r.top));keyboardLayout=[AsdDesktop]::GetKeyboardLayout($thread).ToInt64();text=($texts -join "`n");tree=$tree}
}
function Capture($label) {
 if($label -notmatch '^[a-zA-Z0-9_-]+$'){throw 'Invalid capture label'}
 $p=Guard;$r=New-Object AsdDesktop+RECT;[AsdDesktop]::GetWindowRect($p.MainWindowHandle,[ref]$r)|Out-Null
 $rect=[Drawing.Rectangle]::FromLTRB($r.left,$r.top,$r.right,$r.bottom)
 $visible=[Drawing.Rectangle]::Intersect($rect,[Windows.Forms.SystemInformation]::VirtualScreen)
 if($visible.Width -lt 100 -or $visible.Height -lt 100){throw 'Owned window outside display'}
 $bmp=[Drawing.Bitmap]::new($visible.Width,$visible.Height);$g=[Drawing.Graphics]::FromImage($bmp)
 try{$g.CopyFromScreen($visible.Location,[Drawing.Point]::Empty,$visible.Size);$bmp.Save((Join-Path $out "$label.png"),[Drawing.Imaging.ImageFormat]::Png)}finally{$g.Dispose();$bmp.Dispose()}
 return State
}
try {
 $script:terminal=(Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse | Select-Object -First 1).FullName
 $existing=@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue | Where-Object {$_.Path -eq $script:terminal -and $_.MainWindowHandle -ne 0})
 if($existing.Count -ne 0){throw 'Existing private terminal present; refusing to reuse it'}
 $p=Start-Process -FilePath $script:terminal -ArgumentList '-w new' -PassThru
 $limit=[DateTime]::UtcNow.AddSeconds(20)
 do{Start-Sleep -Milliseconds 100;$matches=@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue | Where-Object {$_.Path -eq $script:terminal -and $_.MainWindowHandle -ne 0})}while($matches.Count -eq 0 -and [DateTime]::UtcNow -lt $limit)
 if($matches.Count -ne 1){throw 'Cannot uniquely identify private terminal'}
 $script:ownedId=$matches[0].Id
 $script:ownedFiletime=$matches[0].StartTime.ToUniversalTime().ToFileTimeUtc()
 [AsdDesktop]::SetForegroundWindow($matches[0].MainWindowHandle)|Out-Null
 Start-Sleep -Milliseconds 500
 $ready=Capture 'startup'
 $ready | ConvertTo-Json -Depth 7 | Set-Content (Join-Path $out 'ready.json') -Encoding UTF8
 $until=[DateTime]::UtcNow.AddMinutes(20)
 while([DateTime]::UtcNow -lt $until){
  $file=Get-ChildItem (Join-Path $out 'commands') -Filter '*.json' | Sort-Object Name | Select-Object -First 1
  if(!$file){Start-Sleep -Milliseconds 100;continue}
  $c=Get-Content $file.FullName -Raw -Encoding UTF8 | ConvertFrom-Json
  $result=@{utc=[DateTime]::UtcNow.ToString('o');command=$c;ok=$true}
  try {
   $p=Guard
   switch($c.kind){
    'text' {foreach($k in @(16,17,18,91,92)){if(([AsdDesktop]::GetAsyncKeyState($k)-band 32768)-ne 0){throw 'Modifier held'}};foreach($ch in $c.text.ToCharArray()){Guard|Out-Null;if([AsdDesktop]::Text([string]$ch)-ne 2){throw 'Text input failed'};Start-Sleep -Milliseconds 20}}
    'paste' {$saved=[Windows.Forms.Clipboard]::GetDataObject();try{[Windows.Forms.Clipboard]::SetText($c.text);Guard|Out-Null;[Windows.Forms.SendKeys]::SendWait('^+v');Start-Sleep -Milliseconds 500}finally{if($saved){[Windows.Forms.Clipboard]::SetDataObject($saved,$true)}else{[Windows.Forms.Clipboard]::Clear()}}}
    'keys' {[Windows.Forms.SendKeys]::SendWait($c.keys)}
    'physical' {foreach($ch in $c.text.ToUpperInvariant().ToCharArray()){Guard|Out-Null;if([AsdDesktop]::Key([uint16][char]$ch)-ne 2){throw 'Physical key input failed'};Start-Sleep -Milliseconds 120}}
    'resize' {[AsdDesktop]::MoveWindow($p.MainWindowHandle,40,40,[int]$c.width,[int]$c.height,$true)|Out-Null}
    'clickListRow' {
     foreach($k in @(1,16,17,18,91,92)){if(([AsdDesktop]::GetAsyncKeyState($k)-band 32768)-ne 0){throw 'Button or modifier already held'}}
     if($c.title -notin @('Density Shell A','Density Shell B')){throw 'Unknown synthetic row title'}
     $element=[Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
     $items=$element.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
     $rectangles=$null;$lines=$null
     foreach($item in $items){$pattern=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){$text=$pattern.DocumentRange.GetText(-1);if($text.StartsWith('Agent Session Desk')){$rectangles=$pattern.DocumentRange.GetBoundingRectangles();$lines=$text -split "`r?`n";break}}}
     if(!$rectangles -or !$lines){throw 'App text rectangles unavailable'}
     $row=-1
     for($i=0;$i -lt $lines.Length;$i++){if($lines[$i].StartsWith('  '+$c.title) -or $lines[$i].StartsWith('▶ '+$c.title) -or $lines[$i].StartsWith('┃ '+$c.title)){$row=$i;break}}
     if($row -ge 0 -and $c.metadata){$row++;if(!$lines[$row].StartsWith('  ╰ ') -and !$lines[$row].StartsWith('  │ ')){throw 'Metadata line not observed'}}
     if($row -lt 0 -or $row -ge $rectangles.Length){throw 'Synthetic list row not observed'}
     $r=$rectangles[$row];$cols=$lines[$row].Length+[regex]::Matches($lines[$row],'[\u4e00-\u9fff]').Count
     if($cols -lt 10 -or $r.Width -lt 50){throw 'Unexpected row width'}
     $x=[int]($r.Left+3.5*$r.Width/$cols);$y=[int]($r.Top+$r.Height/2)
     $count=1;if($c.double){$count=2}
     for($i=0;$i -lt $count;$i++){Guard|Out-Null;[AsdDesktop]::SetCursorPos($x,$y)|Out-Null;if([AsdDesktop]::Click()-ne 2){throw 'Mouse click failed'};if($i+1 -lt $count){Start-Sleep -Milliseconds 120}}
     $result.observedRow=$row;$result.observedColumns=$cols;$result.observedRectangle=@($r.Left,$r.Top,$r.Width,$r.Height);$result.inputClicks=$count
    }
    'click' {$r=New-Object AsdDesktop+RECT;[AsdDesktop]::GetWindowRect($p.MainWindowHandle,[ref]$r)|Out-Null;if($c.x -lt 8 -or $c.y -lt 8 -or $c.x -ge ($r.right-$r.left-8) -or $c.y -ge ($r.bottom-$r.top-8)){throw 'Click outside owned window'};[AsdDesktop]::SetCursorPos($r.left+[int]$c.x,$r.top+[int]$c.y)|Out-Null;[AsdDesktop]::Click()|Out-Null}
    'capture' {$result.state=Capture $c.label}
    'state' {$result.state=State}
    'finish' {$result.state=State;$p.CloseMainWindow()|Out-Null;$result.closed=$true}
    default {throw 'Unknown command'}
   }
  }catch{$result.ok=$false;$result.error=$_.Exception.Message}
  $target=Join-Path (Join-Path $out 'responses') $file.Name
  $result | ConvertTo-Json -Depth 10 | Set-Content $target -Encoding UTF8
  Move-Item $file.FullName (Join-Path (Join-Path $out 'done') $file.Name)
  if($c.kind -eq 'finish'){break}
 }
}catch{@{utc=[DateTime]::UtcNow.ToString('o');error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 4|Set-Content (Join-Path $out 'driver-error.json') -Encoding UTF8}
