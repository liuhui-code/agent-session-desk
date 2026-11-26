$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$old=Join-Path $package 'asd-integrated-74a86959.exe'
$ps=@(Get-Process -Name 'asd-integrated-74a86959' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $old})
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$windows=@([DeskFlow]::HostWindows($hostInfo.terminal.path))
$rows=@()
foreach($p in $ps){$null=$p.Handle;$hosts=0;$bound=0;$nativeTabs=@();$foreground=$false
 foreach($w in $windows){[uint32]$wp=0;$null=[DeskFlow]::GetWindowThreadProcessId($w,[ref]$wp);$current=$p.Id;$found=$false
  for($i=0;$i -lt 32 -and $current -gt 0;$i++){if($current -eq $wp){$found=$true;break};$x=Get-CimInstance Win32_Process -Filter ('ProcessId='+$current) -Property ProcessId,ParentProcessId;if(!$x){break};$current=$x.ParentProcessId}
  if($found){$hosts++;$bound+=[DeskFlow]::OwnerBindings($w);$foreground=$foreground-or ([DeskFlow]::GetForegroundWindow()-eq $w);$root=[Windows.Automation.AutomationElement]::FromHandle($w);$tabs=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem));$nativeTabs+=$tabs.Count}
 }
 $rows+=@{pet_count=@([DeskFlow]::Find('ASD Blocking Reminders',$p.Id)).Count;host_ancestor_windows=$hosts;bound_windows=$bound;native_tab_counts=$nativeTabs;foreground_host=$foreground;age_seconds=[int]([DateTime]::UtcNow-$p.StartTime.ToUniversalTime()).TotalSeconds}
}
@{process_count=$ps.Count;host_window_count=$windows.Count;process_roles=$rows;session_content_exported=0;ui_actions=0;clipboard_calls=0}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $base 'pet-fix-role-probe-20261126.json') -Encoding utf8
