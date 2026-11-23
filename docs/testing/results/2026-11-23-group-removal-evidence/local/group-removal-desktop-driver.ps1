$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$out=Join-Path $base 'group-removal-desktop-20261123/desktop'
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
 public static uint Wheel(int delta){var a=new INPUT[1];a[0].data.mouse.flags=0x800;a[0].data.mouse.data=unchecked((uint)delta);return SendInput(1,a,Marshal.SizeOf(typeof(INPUT)));}
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
 $texts=@();$tree=@();$fontNames=@()
 foreach($el in $items){$name=[string]$el.Current.Name;$tree+=@{name=$name;type=$el.Current.ControlType.ProgrammaticName};$pat=$null;if($el.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pat)){$texts+=,$pat.DocumentRange.GetText(-1);$fontNames+=,[string]$pat.DocumentRange.GetAttributeValue([Windows.Automation.TextPattern]::FontNameAttribute)}}
 $f=[uint32]0;$thread=[AsdDesktop]::GetWindowThreadProcessId([AsdDesktop]::GetForegroundWindow(),[ref]$f)
 return @{caps_lock=[Windows.Forms.Control]::IsKeyLocked([Windows.Forms.Keys]::CapsLock);pid=$p.Id;filetime=$script:ownedFiletime;path=$p.Path;session=$p.SessionId;rect=@($r.left,$r.top,($r.right-$r.left),($r.bottom-$r.top));keyboardLayout=[AsdDesktop]::GetKeyboardLayout($thread).ToInt64();text=($texts -join "`n");tree=$tree;fontNames=@($fontNames|Select-Object -Unique)}
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
$capsBefore=[Windows.Forms.Control]::IsKeyLocked([Windows.Forms.Keys]::CapsLock)
$savedClipboard=$null;$originalClipboardText=$null;$hadText=$false;$savedFormats=@()
try {
 $original=[Windows.Forms.Clipboard]::GetDataObject()
 if($original){
  $savedClipboard=[Windows.Forms.DataObject]::new()
  foreach($format in $original.GetFormats($false)){$data=$original.GetData($format,$false);if($data -is [IO.MemoryStream]){$data=[IO.MemoryStream]::new($data.ToArray())};$savedClipboard.SetData($format,$false,$data);$savedFormats+=,$format}
  $hadText=[Windows.Forms.Clipboard]::ContainsText();if($hadText){$originalClipboardText=[Windows.Forms.Clipboard]::GetText()}
 }
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
 $until=[DateTime]::UtcNow.AddMinutes(40)
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
    'clickNavRow' {
     foreach($k in @(1,16,17,18,91,92)){if(([AsdDesktop]::GetAsyncKeyState($k)-band 32768)-ne 0){throw 'Button or modifier already held'}}
     if($c.title -notin @('Copy A','Copy B','Group C','我的会话','项目 甲','审核')){throw 'Unknown synthetic row title'}
     $element=[Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
     $items=$element.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
     $rectangles=$null;$lines=$null
     foreach($item in $items){$pattern=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){$text=$pattern.DocumentRange.GetText(-1);if($text.StartsWith('Agent Session Desk')){$rectangles=$pattern.DocumentRange.GetBoundingRectangles();$lines=$text -split "`r?`n";break}}}
     if(!$rectangles -or !$lines){throw 'App text rectangles unavailable'}
     $row=-1
     for($i=2;$i -lt $lines.Length;$i++){if(($c.title -in @('我的会话','项目 甲','审核') -and $lines[$i] -match ('[▾▸] '+[regex]::Escape($c.title)+'(?:\s|$)')) -or ($c.title -notin @('我的会话','项目 甲','审核') -and $lines[$i] -match ('^[ ▶┃]+'+[regex]::Escape($c.title)+'(?:\s|$)'))){$row=$i;break}}
     if($row -ge 0 -and $c.metadata){$row++;if(!$lines[$row].TrimStart().StartsWith('╰ ') -and !$lines[$row].TrimStart().StartsWith('│ ')){throw 'Metadata line not observed'}}
     if($row -lt 0 -or $row -ge $rectangles.Length){throw 'Synthetic list row not observed'}
     $r=$rectangles[$row];$cols=$lines[$row].Length+[regex]::Matches($lines[$row],'[\u4e00-\u9fff]').Count
     if($cols -lt 10 -or $r.Width -lt 50){throw 'Unexpected row width'}
     $x=[int]($r.Left+3.5*$r.Width/$cols);$y=[int]($r.Top+$r.Height/2)
     $count=1;if($c.double){$count=2}
     if($c.wheel){if([int]$c.wheel -notin @(-120,120)){throw 'Invalid test wheel delta'};Guard|Out-Null;[AsdDesktop]::SetCursorPos($x,$y)|Out-Null;if([AsdDesktop]::Wheel([int]$c.wheel)-ne 1){throw 'Mouse wheel failed'};$count=0}else{for($i=0;$i -lt $count;$i++){Guard|Out-Null;[AsdDesktop]::SetCursorPos($x,$y)|Out-Null;if([AsdDesktop]::Click()-ne 2){throw 'Mouse click failed'};if($i+1 -lt $count){Start-Sleep -Milliseconds 120}}}
     $result.observedRow=$row;$result.observedColumns=$cols;$result.observedRectangle=@($r.Left,$r.Top,$r.Width,$r.Height);$result.inputClicks=$count
    }
    'dragTab' {
     foreach($key in @(1,16,17,18,91,92)){if(([AsdDesktop]::GetAsyncKeyState($key)-band 32768)-ne 0){throw 'Button or modifier already held'}}
     if($c.source -notin @('Copy A','Copy B') -or $c.destination -notin @('Copy A','Copy B') -or $c.source -eq $c.destination){throw 'Unknown synthetic drag target'}
     $element=[Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
     $items=$element.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
     $rectangles=$null;$lines=$null
     foreach($item in $items){$pattern=$null;if($item.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){$text=$pattern.DocumentRange.GetText(-1);if($text.StartsWith('Agent Session Desk')){$rectangles=$pattern.DocumentRange.GetBoundingRectangles();$lines=$text -split "`r?`n";break}}}
     if(!$rectangles -or !$lines -or $rectangles.Length -lt 2){throw 'Tab row rectangles unavailable'}
     $row=$lines[1];$from=$row.IndexOf([string]$c.source);$to=$row.IndexOf([string]$c.destination)
     if($from -lt 0 -or $to -lt 0){throw 'Both synthetic Tab titles must be visible'}
     $cols=$row.Length;$rectangle=$rectangles[1];$cell=$rectangle.Width/$cols;$y=[int]($rectangle.Top+$rectangle.Height/2)
     $start=[int]($rectangle.Left+($from+0.5)*$cell);$end=[int]($rectangle.Left+($to+0.5)*$cell)
     $pressed=$false
     try{
      Guard|Out-Null;[AsdDesktop]::SetCursorPos($start,$y)|Out-Null;if([AsdDesktop]::MouseButton($true)-ne 1){throw 'Drag press failed'};$pressed=$true;Start-Sleep -Milliseconds 120
      for($step=1;$step -le 12;$step++){Guard|Out-Null;$x=[int]($start+($end-$start)*$step/12);[AsdDesktop]::SetCursorPos($x,$y)|Out-Null;Start-Sleep -Milliseconds 30}
     }finally{if($pressed){[AsdDesktop]::MouseButton($false)|Out-Null}}
     $result.sourceColumn=$from;$result.destinationColumn=$to;$result.tabColumns=$cols;$result.rowRectangle=@($rectangle.Left,$rectangle.Top,$rectangle.Width,$rectangle.Height)
    }
    'expectClipboard' {
     if($c.expected -ne "SCROLL_NEEDLE_中文_α`n"){throw 'Only the synthetic selection may be inspected'}
     $text=[Windows.Forms.Clipboard]::GetText();$normalized=$text.Replace("`r`n","`n")
     if($normalized -cne $c.expected){throw 'Clipboard does not match the synthetic selection'}
     $result.synthetic_text=$normalized;$result.original_length=$text.Length
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

finally{
 if($savedClipboard){[Windows.Forms.Clipboard]::SetDataObject($savedClipboard,$true,10,100)}else{[Windows.Forms.Clipboard]::Clear()}
 $textRestored=(!$hadText -or [Windows.Forms.Clipboard]::GetText() -ceq $originalClipboardText)
 $after=[Windows.Forms.Clipboard]::GetDataObject();$afterFormats=@();if($after){$afterFormats=@($after.GetFormats($false))};$formatRestored=([string]::Join([string][char]0,@($savedFormats|Sort-Object)) -ceq [string]::Join([string][char]0,@($afterFormats|Sort-Object)))
 $restored=$textRestored -and $formatRestored
 @{utc=[DateTime]::UtcNow.ToString('o');caps_before=$capsBefore;caps_after=[Windows.Forms.Control]::IsKeyLocked([Windows.Forms.Keys]::CapsLock);restored=$restored;text_restored=$textRestored;format_names_restored=$formatRestored;nontext_content_verified=$false;original_text_logged=$false}|ConvertTo-Json|Set-Content (Join-Path $out 'clipboard-restored.json') -Encoding UTF8
}
