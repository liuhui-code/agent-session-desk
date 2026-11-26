$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$target=Join-Path $base 'asd-integrated-r1.exe'
$managers=@(Get-Process -Name 'asd-integrated-r1' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target})
foreach($p in $managers){$null=$p.Handle}
$ownWindows=@([DeskFlow]::Find('ASD Integrated Flow R1'))
$diagnostics=@()
foreach($window in $ownWindows){
 $root=[Windows.Automation.AutomationElement]::FromHandle($window)
 $tabs=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem))
 if($tabs.Count -gt 1){throw 'Private test window has unexpected extra tabs; not closed'}
 $text='';$nodes=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
 foreach($node in $nodes){$pattern=$null;if($node.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){$candidate=$pattern.DocumentRange.GetText(65536);if($candidate.Length -gt $text.Length){$text=$candidate}}}
 $diagnostics+=@{review_reference_present=$text.Contains('review.md');open_failed=$text.Contains('Could not open Markdown');open_requested=$text.Contains('Markdown open requested');connection_read_only=$text.Contains('Read-only last');owner_binding_present=([DeskFlow]::OwnerBindings($window)-gt 0)}
}
$ids=@($managers|ForEach-Object{$_.Id})
$notes=@(Get-CimInstance Win32_Process -Property ProcessId,ParentProcessId,ExecutablePath -Filter "Name='notepad.exe'"|Where-Object{$_.ParentProcessId -in $ids})
foreach($n in $notes){$p=Get-Process -Id $n.ProcessId;$null=$p.Handle;foreach($window in @([DeskFlow]::Find('',$p.Id))){$null=[DeskFlow]::SendMessage($window,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)};$null=$p.WaitForExit(10000)}
foreach($window in $ownWindows){$null=[DeskFlow]::SendMessage($window,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)}
$ends=@($managers|ForEach-Object{if(!$_.HasExited){$null=$_.WaitForExit(15000)};$_.HasExited})
@{owned_windows=$ownWindows.Count;owned_manager_count=$managers.Count;owned_notepad_count=$notes.Count;manager_ends=$ends;controlled_surface_checks=$diagnostics;clipboard_calls=0;forced_process_ends=0}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $base 'integrated-r1-cleanup.json') -Encoding utf8
