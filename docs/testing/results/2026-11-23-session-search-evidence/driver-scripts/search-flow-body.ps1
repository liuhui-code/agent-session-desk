function SearchGeometry([string]$label,[int]$expectedColumns){
 $r=InvokeVisual @{kind='state'};$lines=$r.state.text -split '\r?\n';$border=@($lines|Where-Object{$_ -match '^\s*╭─+╮\s*$'})
 if($border.Count -ne 2){throw 'Expected outer and input search borders'}
 $geometry=@{label=$label;columns=$border[0].Length;outer_x=$border[0].IndexOf('╭');outer_width=($border[0].IndexOf('╮')-$border[0].IndexOf('╭')+1);input_x=$border[1].IndexOf('╭');input_width=($border[1].IndexOf('╮')-$border[1].IndexOf('╭')+1)}
 $expected=@{120=@(62,52);63=@(55,45);49=@(41,31)}[$expectedColumns]
 if($geometry.columns -ne $expectedColumns -or $geometry.outer_width -ne $expected[0] -or $geometry.input_width -ne $expected[1]){throw 'Search geometry differs from pinned-source width contract'}
 $script:geometry.Add($geometry);$script:geometry.ToArray()|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'search-geometry.json') -Encoding utf8
}
function CheckDraft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title
 if($o.capture_text -notmatch [regex]::Escape($draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw 'Original draft was changed or submitted before explicit Enter'}
}
$script:geometry=[Collections.Generic.List[object]]::new()
try{
 $ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20);while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $ready)){throw 'Driver not ready'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A';RememberUI;$ids=@(CheckShells)
 Text 'Write-Output (''A:''+$PID) # search-draft'
 Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '本地会话搜索';$null=WaitText '3 个结果'
 SearchGeometry 'empty-120' 120;$null=Shot 'search-empty-120-public'
 Key '^{PGDN}';Key '{F6}';$null=WaitText '本地会话搜索';Text 'q';$null=WaitText '无匹配会话'
 Key '{ESC}';$null=WaitText '管理模式';CheckDraft 'a-after-cancel' 'Copy A' 'search-draft' 'A'
 Text '/';$null=WaitText '3 个结果';Paste '  COPY b  ';$null=WaitText '› Copy B';Key '{ENTER}';$null=WaitText 'COPY_READY Copy B'
 $b=Observe 'b-opened-by-search' 'Copy B';$nonempty=@(($b.capture_text -split '\r?\n')|Where-Object{$_.Trim().Length -gt 0});if($nonempty[-1].Trim() -cne '>'){throw 'Search Enter or query reached B terminal'}
 Text 'Write-Output (''B:''+$PID) # b-search-draft'
 Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '本地会话搜索';Paste '中文';$null=WaitText '› 中文会话';$null=WaitText '1 个结果'
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '› 中文会话';SearchGeometry 'unicode-63' 63;$null=Shot 'search-unicode-63-public'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '› 中文会话';SearchGeometry 'unicode-49' 49;$null=Shot 'search-unicode-49-public'
 Key '{ENTER}';$null=WaitText 'COPY_READY 中文会话'
 Text 'Write-Output (''C:''+$PID) # third-draft'
 Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '本地会话搜索';Paste 'zzzzzzzzz';$null=WaitText '无匹配会话';Key '{ENTER}';$null=WaitText '无匹配会话';$null=Shot 'search-no-results-49-public'
 Key '{ESC}';$null=WaitText '管理模式';CheckDraft 'c-after-empty-enter' '中文会话' 'third-draft' 'C';CheckDraft 'b-after-unicode-search' 'Copy B' 'b-search-draft' 'B'
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY 中文会话';Key '{F6}';$null=WaitText 'third-draft';Key '{ENTER}';$text=WaitText 'C:(\d+) *│';$cPid=[int]([regex]::Match($text,'C:(\d+) *│').Groups[1].Value)
 Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '3 个结果';Paste 'Copy B';$null=WaitText '› Copy B';Key '{ENTER}';$null=WaitText 'b-search-draft';Key '{ENTER}';$text=WaitText 'B:(\d+) *│';$bPid=[int]([regex]::Match($text,'B:(\d+) *│').Groups[1].Value)
 Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '3 个结果';Paste 'Copy A';$null=WaitText '› Copy A';Key '{ENTER}';$null=WaitText 'search-draft';Key '{ENTER}';$text=WaitText 'A:(\d+) *│';$aPid=[int]([regex]::Match($text,'A:(\d+) *│').Groups[1].Value)
 if($aPid -ne $ids[0].pid -or $bPid -ne $ids[1].pid -or $cPid -ne $ids[2].pid){throw 'Explicit commands did not execute in original shells'}
 $null=Observe 'a-explicit-enter' 'Copy A';$null=Observe 'b-explicit-enter' 'Copy B';$null=Observe 'c-explicit-enter' '中文会话';$null=Shot 'search-original-a-returned-public'
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit
 $command="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $command;Key '{ENTER}';$null=WaitText 'COPY_READY Copy A';RememberUI;$null=CheckShells
 Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '3 个结果';SearchGeometry 'reopened-empty-120' 120;$null=Shot 'search-reopened-empty-120-public'
 Key '{ESC}';$null=WaitText '管理模式';Text 'q';WaitExit
 $after=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $after.Count -ne 3){throw 'Original three managed sessions not preserved'}
 foreach($s in $after){if(@($sessions|Where-Object{$_.session_id -eq $s.session_id -and $_.run_id -eq $s.run_id}).Count -ne 1){throw 'Original Session/Run changed'}}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';commands=$script:sequence;before=$sessions;after=$after;shell_processes=@(CheckShells);ui_processes=$script:uis.ToArray();geometry=$script:geometry.ToArray();native_observations=$script:observations.Count;checks=@('search query stays outside native input','case/whitespace ranked exact target','Chinese title search','120/63/49 source-derived widths','no-results Enter stays in modal','cancel preserves original drafts','explicit Enter executes A/B/C in original roots','UI reopen starts fresh query','Session/Run and PID/FILETIME preserved')}|ConvertTo-Json -Depth 9|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8;throw}
