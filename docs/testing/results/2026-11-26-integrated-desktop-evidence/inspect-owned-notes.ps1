$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$wanted=@((Join-Path $base 'integrated-desktop-run1-20261126/review.md'),(Join-Path $base 'integrated-desktop-run2-20261126/review.md'))
$notes=@(Get-CimInstance Win32_Process -Filter "Name='notepad.exe'" -Property ProcessId,ExecutablePath,CommandLine|Where-Object{$line=$_.CommandLine;[bool]@($wanted|Where-Object{$line -and $line.IndexOf($_,[StringComparison]::OrdinalIgnoreCase)-ge 0}).Count})
$results=@()
foreach($item in $notes){
 $p=Get-Process -Id $item.ProcessId;$null=$p.Handle
 $owned=@([DeskFlow]::Find('',$p.Id));$seen=$false
 foreach($window in $owned){
  $root=[Windows.Automation.AutomationElement]::FromHandle($window)
  $nodes=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.Condition]::TrueCondition)
  foreach($node in $nodes){$pattern=$null;if($node.TryGetCurrentPattern([Windows.Automation.TextPattern]::Pattern,[ref]$pattern)){if($pattern.DocumentRange.GetText(1024).Contains('ASD_DESKTOP_DOC_v1')){$seen=$true}}}
 }
 $isSystem=(Get-FileHash $p.Path).Hash -ceq (Get-FileHash (Join-Path ([Environment]::SystemDirectory) 'notepad.exe')).Hash
 if($seen -and $isSystem){foreach($window in $owned){$null=[DeskFlow]::SendMessage($window,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)}}
 $results+=@{controlled_file_path_matched=$true;controlled_document_visible=$seen;system_notepad_binary=$isSystem;closed_normally=($seen -and $p.WaitForExit(10000))}
}
@{owned_document_windows=$results;matched_count=$notes.Count;command_lines_exported=0;other_document_content_exported=0;clipboard_calls=0}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $base 'integrated-owned-notes.json') -Encoding utf8
