$helper=Join-Path $base 'read-group-metadata.exe'
if((Get-FileHash $helper).Hash.ToLowerInvariant() -cne '15d1395962ec77a44f0f7768acc83461f47804b0d2f4fd7711b3b59800c746ce'){throw 'Read-only observer differs'}
$script:metadata=[Collections.Generic.List[object]]::new();$script:preferences=[Collections.Generic.List[object]]::new()
function Meta([string]$label,[string]$path=''){
 if(!$path){$path=Join-Path $testRoot 'state/metadata.sqlite3'}
 $v=(& $helper $path|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $v.Version -ne 4){throw 'Read-only metadata observation failed'}
 $script:metadata.Add(@{label=$label;value=$v});$script:metadata.ToArray()|ConvertTo-Json -Depth 12|Set-Content (Join-Path $testRoot 'metadata-observations.json') -Encoding utf8
 return $v
}
function Pref([string]$label,[string]$density,[string]$folds){
 $path=Join-Path $testRoot 'state/ui.toml';$until=[DateTime]::UtcNow.AddSeconds(8)
 do{$text=ReadSharedText $path;$ok=$text -match '(?m)^schema_version = 3$' -and $text -match ('(?m)^sidebar_density = "'+$density+'"$') -and $text -match '(?m)^sidebar_mode = "grouped"$' -and $text -match ('(?m)^collapsed_groups = "'+$folds+'"$');if($ok){$script:preferences.Add(@{label=$label;density=$density;folds=$folds;text=$text;sha256=(Get-FileHash $path).Hash});$script:preferences.ToArray()|ConvertTo-Json -Depth 6|Set-Content (Join-Path $testRoot 'preferences.json') -Encoding utf8;return};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
 throw ('Preference not observed: '+$label)
}
function ClickNav([string]$title,[bool]$metadata=$false){$null=InvokeVisual @{kind='clickNavRow';title=$title;metadata=$metadata}}
function Draft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title;$lines=@(($o.capture_text -split '\r?\n')|Where-Object{$_.Trim()})
 if(!$lines.Count -or $lines[-1].Trim() -cne ('> '+$draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw ('Pending original draft differs: '+$label)}
}
function ViewCols([int]$cols){
 $text=StateText;$lines=@(($text -split '\r?\n')|Where-Object{$_ -match 'Copy A ×' -and $_ -match 'Copy B ×'});if($lines.Count -ne 1 -or $lines[0].Length -ne $cols){throw ('Unique Tab row columns differ from '+$cols)}
}
function SameOriginals($rows){foreach($s in $originalSessions){if(@($rows|Where-Object{$_.session_id -ceq $s.session_id -and $_.run_id -ceq $s.run_id -and $_.backend_id -ceq $s.backend_id -and $_.command -ceq $s.command}).Count -ne 1){throw 'Original Session/Run/Backend changed'}}}
function CurrentSessions{$rows=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0){throw 'Actual discovery failed'};SameOriginals $rows;return $rows}
function Identity([string]$name){$i=(ReadSharedText (Join-Path $testRoot ('work/'+$name.Replace(' ','_')+'.identity.json')))|ConvertFrom-Json;$p=Get-Process -Id $i.pid;if($p.Path -ne $i.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $i.filetime){throw 'Original synthetic process changed'};return $i}
$a="Write-Output ('GROUP_A:'+`$PID)";$b="Write-Output ('GROUP_B:'+`$PID)";$c="Write-Output ('GROUP_C:'+`$PID)"
$rootDefault=[IO.Path]::GetFullPath((Join-Path $testRoot 'work/root-default'));$reviewDefault=[IO.Path]::GetFullPath((Join-Path $testRoot 'work/review-default'))
$originalSessions=@($sessions);$before=@();$rootGroup=$null;$childGroup=$null
try{
 $ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20);while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $ready)){throw 'Desktop driver not ready'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A';RememberUI;$before=@(CheckShells)
 $seed=(Get-FileHash (Join-Path $testRoot 'ui-v1-before.toml')).Hash
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $seed){throw 'Startup rewrote legacy preferences'}
 Paste $a;Draft 'A-initial' 'Copy A' $a 'GROUP_A';Key '^{PGDN}';$null=WaitText 'COPY_READY Copy B';Paste $b;Draft 'B-initial' 'Copy B' $b 'GROUP_B';Key '^{PGUP}';$null=WaitText ([regex]::Escape($a));Key '{F6}';$null=WaitText '管理模式';ViewCols 120
 Text 'g';$null=WaitText '创建分组';$null=Shot 'create-root-empty-120-public'
 Text '项目 甲';Key '{TAB}';Paste $rootDefault;$null=Shot 'create-root-filled-120-public'
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '创建分组';$null=Shot 'create-root-63-public'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '创建分组';$null=Shot 'create-root-49-public'
 $null=InvokeVisual @{kind='resize';width=1113;height=657};Key '{ENTER}';$null=WaitText '[▾▸] 项目 甲 \(0\)'
 $m=Meta 'root-created';$rootGroup=@($m.Groups|Where-Object{$_.Name -ceq '项目 甲'})[0]
 if(!$rootGroup -or $rootGroup.ParentID -ne '' -or $rootGroup.DefaultDir -cne $rootDefault -or @($m.Groups).Count -ne 2){throw 'Root group persisted differently'}
 Text 'g';$null=WaitText '创建子组';$null=Shot 'create-subgroup-empty-120-public'
 Text '审核';Key '{TAB}';Paste $reviewDefault;Key '{UP}';$null=WaitText '创建分组';Key '{DOWN}';$null=WaitText '创建子组';Key '{ENTER}';$null=WaitText '[▾▸] 审核 \(0\)'
 $m=Meta 'subgroup-created';$childGroup=@($m.Groups|Where-Object{$_.Name -ceq '审核'})[0]
 if(!$childGroup -or $childGroup.ParentID -cne $rootGroup.ID -or $childGroup.DefaultDir -cne $reviewDefault -or @($m.Groups).Count -ne 3){throw 'Nested group persisted differently'}
 Key '{HOME}';Key '{DOWN}';Text 'M';$null=WaitText '移动到分组';$null=Shot 'move-dialog-120-public'
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=Shot 'move-dialog-63-public';$null=InvokeVisual @{kind='resize';width=400;height=539};$null=Shot 'move-dialog-49-public';$null=InvokeVisual @{kind='resize';width=1113;height=657}
 Key '{DOWN}';Key '{DOWN}';Key '{ENTER}';$null=WaitText '[▾▸] 审核 \(1\)'
 $m=Meta 'A-moved-to-child';$member=@($m.Members|Where-Object{$_.SessionID -ceq $originalSessions[0].session_id})[0]
 if(!$member -or $member.GroupID -cne $childGroup.ID -or $member.Revision -ne 1){throw 'A membership differs'}
 ClickNav 'Copy B' $true;Text 'M';$null=WaitText '移动到分组';Key '{DOWN}';Key '{ENTER}';$null=WaitText '[▾▸] 项目 甲 \(1\)';$null=WaitText '我的会话 \(0\)'
 $m=Meta 'B-moved-to-root';$member=@($m.Members|Where-Object{$_.SessionID -ceq $originalSessions[1].session_id})[0]
 if(!$member -or $member.GroupID -cne $rootGroup.ID -or $member.Revision -ne 1 -or @($m.Runs).Count -ne 2){throw 'B membership or Run count differs'}
 $live=@(CurrentSessions);if($live.Count -ne 2){throw 'Group operation started an extra Run'};Draft 'A-after-moves' 'Copy A' $a 'GROUP_A';Draft 'B-after-moves' 'Copy B' $b 'GROUP_B';$null=Shot 'tree-two-members-120-public'
 Key '2';Key '{LEFT}';$null=WaitText '▸ 项目 甲';Pref 'root-folded' 'compact' $rootGroup.ID;$foldHash=(Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash
 Key '^{PGDN}';$null=WaitText '▾ 项目 甲';Draft 'B-tab-reveal' 'Copy B' $b 'GROUP_B'
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $foldHash){throw 'Tab reveal saved expansion'}
 Key '2';Key '{LEFT}';Pref 'before-search-fold' 'compact' $rootGroup.ID;$foldHash=(Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash
 Text '/';$null=WaitText '本地会话搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText '▾ 审核';Draft 'A-search-reveal' 'Copy A' $a 'GROUP_A'
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $foldHash){throw 'Search reveal saved expansion'}
 Key '{F6}';$null=WaitText '管理模式';ClickNav '审核';Text 'n';$null=WaitText '新建终端'
 Text 'Group C';Key '{TAB}';Paste $shell;Key '{TAB}';Key '{END}';$null=WaitText 'review-default';$null=Shot 'new-in-child-default-directory-public';Key '{TAB}'
 $args=@('-NoLogo','-NoProfile','-NoExit','-File',(Join-Path $testRoot 'work/prompt.ps1'),'Group C')|ConvertTo-Json -Compress
 Paste $args;Key '^s';$null=WaitText '终端模式';$null=WaitText 'COPY_READY Group C'
 $current=@(CurrentSessions);if($current.Count -ne 3){throw 'Expected exactly one group-created terminal'};$created=@($current|Where-Object{$_.title -ceq 'Group C'})[0];if(!$created){throw 'Group-created Session not discovered'}
 $script:sessions=@($originalSessions)+@($created);$script:sessions|ConvertTo-Json -Depth 8|Set-Content (Join-Path $testRoot 'sessions.json') -Encoding utf8
 $ci=Identity 'Group C';if($ci.cwd -cne $childGroup.DefaultDir -or $ci.pid -in @($before.pid)){throw 'New terminal default cwd or PID differs'}
 $m=Meta 'group-created-terminal';$member=@($m.Members|Where-Object{$_.SessionID -ceq $created.session_id})[0]
 if(!$member -or $member.GroupID -cne $childGroup.ID -or @($m.Runs).Count -ne 3 -or @($m.Members).Count -ne 3){throw 'New Run membership not committed'}
 Paste $c;Draft 'C-initial' 'Group C' $c 'GROUP_C'
 Key '{F6}';$null=WaitText '管理模式';Text 'S';$null=WaitText '侧栏密度';Key '{LEFT}';Key '{ESC}';Pref 'density-full' 'full' '';$null=Shot 'tree-full-120-public'
 ClickNav 'Copy A' $true;Key '{ENTER}';$null=WaitText '终端模式';Draft 'A-full-metadata-open' 'Copy A' $a 'GROUP_A';Key '{F6}';Text 'S';$null=WaitText '侧栏密度';Key '{RIGHT}';Key '{RIGHT}';Key '{ESC}';Pref 'density-minimal' 'minimal' '';$null=Shot 'tree-minimal-120-public'
 Text 'S';$null=WaitText '侧栏密度';Key '{RIGHT}';Key '{ESC}';Pref 'density-auto' 'auto' ''
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '管理模式';ViewCols 63;$null=Shot 'tree-auto-63-public'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '管理模式';ViewCols 49;$null=Shot 'tree-auto-49-public';ClickNav 'Copy A';Key '{ENTER}';$null=WaitText '终端模式';Draft 'A-narrow-original-draft' 'Copy A' $a 'GROUP_A'
 $null=InvokeVisual @{kind='resize';width=1113;height=657};Key '{F6}';$null=WaitText '管理模式';Text 'S';$null=WaitText '侧栏密度';Key '{LEFT}';Key '{LEFT}';Key '{LEFT}';Key '{ESC}';Pref 'final-full' 'full' '';Key '2';Key '{LEFT}';$null=WaitText '▸ 项目 甲';Pref 'before-reopen-fold' 'full' $rootGroup.ID
 $backup=Join-Path $testRoot 'group-backup.sqlite3';$null=& $product @g metadata backup $backup;if($LASTEXITCODE -ne 0){throw 'Live v4 metadata snapshot failed'};$backupMeta=Meta 'backup-groups' $backup
 if(@($backupMeta.Groups).Count -ne 3 -or @($backupMeta.Members).Count -ne 3){throw 'Backup omitted groups or members'}
 Text 'q';WaitExit;$null=@(CurrentSessions);Draft 'A-after-quit' 'Copy A' $a 'GROUP_A';Draft 'B-after-quit' 'Copy B' $b 'GROUP_B';Draft 'C-after-quit' 'Group C' $c 'GROUP_C'
 $launch="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $launch;Key '{ENTER}';$null=WaitText '终端模式';RememberUI;Key '{F6}';$null=WaitText '管理模式';$null=WaitText '▸ 项目 甲';Pref 'reopen-fold' 'full' $rootGroup.ID;$null=Shot 'tree-folded-reopen-120-public'
 $m=Meta 'reopened';if(@($m.Groups).Count -ne 3 -or @($m.Members).Count -ne 3 -or @($m.Runs).Count -ne 3){throw 'Reopen lost group metadata'}
 Text '/';$null=WaitText '本地会话搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText '终端模式';Draft 'A-reopen-before-enter' 'Copy A' $a 'GROUP_A';Key '{ENTER}';$null=WaitText ('GROUP_A:'+$before[0].pid);$oa=Observe 'A-explicit-enter' 'Copy A';if($oa.capture_text -notmatch ('(?m)^GROUP_A:'+$before[0].pid+'\s*$')){throw 'A executed in another PID'}
 Key '^{PGDN}';$null=WaitText '终端模式';Draft 'B-reopen-before-enter' 'Copy B' $b 'GROUP_B';Key '{ENTER}';$null=WaitText ('GROUP_B:'+$before[1].pid);$ob=Observe 'B-explicit-enter' 'Copy B';if($ob.capture_text -notmatch ('(?m)^GROUP_B:'+$before[1].pid+'\s*$')){throw 'B executed in another PID'}
 Key '^{PGDN}';$null=WaitText '终端模式';Draft 'C-reopen-before-enter' 'Group C' $c 'GROUP_C';Key '{ENTER}';$null=WaitText ('GROUP_C:'+$ci.pid);$oc=Observe 'C-explicit-enter' 'Group C';if($oc.capture_text -notmatch ('(?m)^GROUP_C:'+$ci.pid+'\s*$')){throw 'C executed in another PID'}
 Key '{F6}';$null=WaitText '管理模式';Key '%v';$null=WaitText '会话 平铺';$null=Shot 'flat-three-sessions-public';Key '%v';$null=WaitText '会话 分组'
 Text 'q';WaitExit;$after=@(CurrentSessions);if($after.Count -ne 3){throw 'Final UI exit changed live set'};$null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;before=$originalSessions;after=$after;before_processes=$before;after_processes=@((Identity 'Copy A'),(Identity 'Copy B'),(Identity 'Group C'));ui_processes=$script:uis.ToArray();root_group=$rootGroup;child_group=$childGroup;preference_snapshots=$script:preferences.Count;metadata_snapshots=$script:metadata.Count;scope='Three ordinary shells, actual group creation/move/default directory, snapshots/reopen, 120/63/49 columns and original drafts. Unicode text input is not physical uppercase or IME-composition acceptance.'}|ConvertTo-Json -Depth 12|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8;throw}
