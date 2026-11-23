function RenameGeometry([string]$label,[int]$columns){
 $state=StateText;$lines=$state -split '\r?\n';$top=-1
 for($i=0;$i -lt $lines.Length;$i++){if($lines[$i] -match '╭─+╮'){$top=$i;break}}
 if($top -lt 0){throw 'Rename border not observed'}
 $line=$lines[$top];$x=$line.IndexOf('╭');$width=$line.IndexOf('╮')-$x+1
 $want=46;if($columns -eq 49){$want=41}
 $g=@{label=$label;columns=$line.Length;outer_x=$x;outer_width=$width;top=$top;bottom=$top+9}
 if($line.Length -ne $columns -or $width -ne $want -or $lines[$top+9] -notmatch '╰─+╯' -or $lines[$top+4] -notmatch '> '){throw 'Rename geometry differs from pinned component'}
 $script:geometry.Add($g);$script:geometry.ToArray()|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'rename-geometry.json') -Encoding utf8
}
function CurrentRows {
 $text=(& $product @g list --json|Out-String);if($LASTEXITCODE -ne 0){throw 'List failed'};$rows=$text|ConvertFrom-Json
 if($rows.Count -ne 3){throw 'Original sessions changed count'}
 foreach($row in $rows){if(@($sessions|Where-Object{$_.session_id -eq $row.session_id -and $_.run_id -eq $row.run_id -and $_.backend_id -eq $row.backend_id}).Count -ne 1){throw 'Original Session/Run changed'}}
 return $rows
}
function OpenOriginalBySearch($target){
 Text '/';$null=WaitText '本地会话搜索';Paste 'Copy A';$null=WaitText '2 个结果'
 $id=$target.session_id.Substring(0,6);$selected='› Copy A \['+$id+'\]'
 if((StateText) -notmatch $selected){Key '{DOWN}';$null=WaitText $selected}
 Key '{ENTER}';$null=WaitText '终端模式'
}
$script:geometry=[Collections.Generic.List[object]]::new()
$a=@($sessions|Where-Object{$_.title -ceq 'Copy A'})[0];$b=@($sessions|Where-Object{$_.title -ceq 'Copy B'})[0]
if($a.session_id.Substring(0,6) -eq $b.session_id.Substring(0,6)){throw 'Synthetic short prefixes collide; inspect before flow'}
$draftA='Write-Output (''RENAME_A:''+$PID)';$draftB='Write-Output (''RENAME_B:''+$PID)'
$renamed='评审 ; ''literal'''
try{
 $ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20)
 while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $ready)){throw 'Driver not ready'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A';RememberUI;$beforeProcesses=@(CheckShells)
 Paste $draftA;CheckDraft 'A-before-rename' 'Copy A' $draftA 'RENAME_A'
 Key '{F6}';$null=WaitText '管理模式';Key '{DOWN}';Text 'r';$null=WaitText '重命名会话'
 $null=WaitText '> Copy B';RenameGeometry 'rename-background-120' 120;$null=Shot 'rename-background-120-public'
 Key '^u';Paste $renamed;Key '{ENTER}';$null=WaitText '管理模式';$null=WaitText 'COPY_READY Copy A'
 $rows=@(CurrentRows);$named=@($rows|Where-Object{$_.session_id -eq $b.session_id})[0]
 if($named.title -cne $renamed -or $named.title_revision -ne 1){throw 'Background name not persisted'}
 CheckDraft 'A-after-background-rename' 'Copy A' $draftA 'RENAME_A'
 $f=NativeFrame $b;if(($f.layout.rows_v2.runs.text -join '') -notmatch 'COPY_READY Copy B'){throw 'Original background pane changed'}
 Text '/';$null=WaitText '本地会话搜索';Paste '评审';$null=WaitText '1 个结果';Key '{ENTER}';$null=WaitText 'COPY_READY Copy B';$null=WaitText '终端模式'
 Paste $draftB;CheckDraft 'B-before-duplicate-name' 'Copy B' $draftB 'RENAME_B'
 Key '{F6}';$null=WaitText '管理模式';Text 'r';$null=WaitText '重命名会话';Key '^u';Paste 'Copy A';Key '{ENTER}';$null=WaitText '管理模式'
 $rows=@(CurrentRows);if(@($rows|Where-Object{$_.title -ceq 'Copy A'}).Count -ne 2){throw 'Duplicate name overwrote a Session'}
 $null=WaitText ('Copy A \['+$a.session_id.Substring(0,6)+'\]');$null=WaitText ('Copy A \['+$b.session_id.Substring(0,6)+'\]')
 CheckDraft 'B-after-duplicate-name' 'Copy B' $draftB 'RENAME_B'
 Text 'r';$null=WaitText '重命名会话';$null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '> Copy A';RenameGeometry 'rename-prefill-63' 63;$null=Shot 'rename-prefill-63-public'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '> Copy A';RenameGeometry 'rename-prefill-49' 49;$null=Shot 'rename-prefill-49-public'
 Key '^u';Paste 'cancelled-name';Key '{ESC}';$null=WaitText '管理模式';$null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText '管理模式'
 $rows=@(CurrentRows);if(@($rows|Where-Object{$_.title -ceq 'Copy A'}).Count -ne 2){throw 'Cancel changed persisted names'}
 Key '^r';$null=WaitText '管理模式';$null=Shot 'rename-duplicate-identities-120-public'
 CheckDraft 'A-before-UI-exit' 'Copy A' $draftA 'RENAME_A';CheckDraft 'B-before-UI-exit' 'Copy B' $draftB 'RENAME_B'
 Text 'q';WaitExit
 $command="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Text $command;Key '{ENTER}';$null=WaitText '(?m)^Agent Session Desk';$null=WaitText '终端模式';RememberUI
 Key '{F6}';$null=WaitText '管理模式';OpenOriginalBySearch $a;$null=WaitText 'COPY_READY Copy A'
 CheckDraft 'A-after-UI-reopen' 'Copy A' $draftA 'RENAME_A'
 Key '{ENTER}';$null=WaitText ('RENAME_A:'+$beforeProcesses[0].pid)
 $afterA=Observe 'A-explicit-enter' 'Copy A';if($afterA.capture_text -notmatch ('(?m)^RENAME_A:'+$beforeProcesses[0].pid+'\s*$')){throw 'A explicit Enter reached wrong process'}
 Key '{F6}';$null=WaitText '管理模式';OpenOriginalBySearch $b;$null=WaitText 'COPY_READY Copy B'
 CheckDraft 'B-after-UI-reopen' 'Copy B' $draftB 'RENAME_B'
 Key '{ENTER}';$null=WaitText ('RENAME_B:'+$beforeProcesses[1].pid)
 $afterB=Observe 'B-explicit-enter' 'Copy B';if($afterB.capture_text -notmatch ('(?m)^RENAME_B:'+$beforeProcesses[1].pid+'\s*$')){throw 'B explicit Enter reached wrong process'}
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit
 $after=@(CurrentRows);$afterProcesses=@(CheckShells);$null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;before=$sessions;after=$after;before_processes=$beforeProcesses;after_processes=$afterProcesses;ui_processes=$script:uis.ToArray();geometry=$script:geometry.ToArray();scope='Three ordinary shells, background rename, duplicate labels, cancel, 120/63/49 columns, original drafts/PIDs/FILETIMEs and exact Enter after UI reopen; no real agent or physical IME claim'}|ConvertTo-Json -Depth 10|Set-Content (Join-Path $testRoot 'rename-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'rename-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'rename-failure.json') -Encoding utf8;throw}
