$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'group-order-desktop-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'asd-group-order.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-group-order';$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
$sessions=([IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json);$script:panes=@{};$script:sequence=0;$script:observations=[Collections.Generic.List[object]]::new();$script:uis=[Collections.Generic.List[object]]::new()
function ReadSharedText($path){
 $stream=[IO.File]::Open($path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete))
 try{$reader=[IO.StreamReader]::new($stream,[Text.Encoding]::UTF8,$true);try{return $reader.ReadToEnd()}finally{$reader.Dispose()}}finally{$stream.Dispose()}
}
function ReadResponse($path){
 $until=[DateTime]::UtcNow.AddSeconds(10)
 do{try{$r=(ReadSharedText $path)|ConvertFrom-Json;if($null -ne $r -and $r.PSObject.Properties['ok']){return $r}}catch{};Start-Sleep -Milliseconds 50}while([DateTime]::UtcNow -lt $until)
 throw 'Complete response not yet readable; action will not be replayed'
}
function Action($id,$value){
 $name=('{0:d3}.json' -f $id);$path=Join-Path $root "commands/$name";$temp=Join-Path $root "$name.tmp"
 [IO.File]::WriteAllText($temp,($value|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false));Move-Item $temp $path
 $response=Join-Path $root "responses/$name";$waitSeconds=15;if($value.kind -eq 'text'){$waitSeconds=[Math]::Min(55,15+[Math]::Ceiling(0.05*$value.text.Length))};$until=[DateTime]::UtcNow.AddSeconds($waitSeconds)
 while(!(Test-Path $response) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $response)){throw 'Driver response timeout'};$r=ReadResponse $response
 if(!$r.ok){throw ('Driver refused action: '+$r.error)};return $r
}
function InvokeVisual($value){$script:sequence++;return Action $script:sequence $value}
function Key($keys){$null=InvokeVisual @{kind='keys';keys=$keys}}
function Text($text){$null=InvokeVisual @{kind='text';text=$text}}
function Paste($text){$null=InvokeVisual @{kind='paste';text=$text}}
function Shot($name){$r=InvokeVisual @{kind='capture';label=$name};return $r.state.text}
function StateText{$r=InvokeVisual @{kind='state'};return $r.state.text}
function WaitText($pattern){
 $until=[DateTime]::UtcNow.AddSeconds(12)
 do{$text=StateText;if($text -match $pattern){return $text};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
 throw "Expected screen not observed: $pattern"
}
function WaitExit{
 $until=[DateTime]::UtcNow.AddSeconds(12)
 do{$text=StateText;$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $product});if($text -notmatch '(?m)^Agent Session Desk' -and $live.Count -eq 0){return};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
 throw 'UI has not exited; no shell launch command will be typed'
}
function RememberUI{
 $found=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $product});if($found.Count -ne 1){throw 'UI image not unique'};$p=Get-Process -Id $found[0].ProcessId
 $script:uis.Add(@{pid=$p.Id;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path;session=$p.SessionId})
 $script:uis.ToArray()|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'ui-identities.json') -Encoding utf8
}
function NativeFrame($s){
 if($s -is [Array] -or $s.backend_id -isnot [string]){throw 'Expected a single typed Session identity'}
 $data=Join-Path (Join-Path $testRoot 'runtime') $ns;$reg=$ns+'__'+$s.backend_id
 $port=[int]((ReadSharedText (Join-Path $data ($reg+'.port')))).Trim();$key=((ReadSharedText (Join-Path $data ($reg+'.key')))).Trim()
 $tcp=[Net.Sockets.TcpClient]::new();$tcp.Connect('127.0.0.1',$port)
 try{$stream=$tcp.GetStream();$stream.ReadTimeout=5000;$stream.WriteTimeout=3000;$reader=[IO.StreamReader]::new($stream,[Text.UTF8Encoding]::new($false),$false,4096,$true)
  $b=[Text.Encoding]::UTF8.GetBytes("AUTH $key`ndump-state`n");$stream.Write($b,0,$b.Length);$stream.Flush();if($reader.ReadLine() -ne 'OK'){throw 'Read-only auth failed'}
  $line=$reader.ReadLine();if(!$line -or [Text.Encoding]::UTF8.GetByteCount($line) -gt 2MB){throw 'Invalid read-only frame'};$f=$line|ConvertFrom-Json
  if($f.session_name -ne $s.backend_id -or $f.layout.type -ne 'leaf' -or !$f.layout.active){throw 'Original synthetic Session/pane changed'}
  if($script:panes.ContainsKey($s.session_id)){if($script:panes[$s.session_id] -ne $f.layout.id){throw 'Original pane changed'}}else{$script:panes[$s.session_id]=$f.layout.id}
  return $f
 }finally{$tcp.Dispose()}
}
function Observe([string]$label,[string]$title,[switch]$Alternate){
 $s=@($sessions|Where-Object{$_.title -ceq $title});if($s.Count -ne 1){throw 'Synthetic identity not unique'};$s=$s[0];$f=NativeFrame $s;$p=$f.layout
 $o=@{label=$label;session_id=$s.session_id;run_id=$s.run_id;pane=$p.id;columns=$p.cols;rows=$p.rows;copy_mode=$p.copy_mode;backend_alternate_hint=$p.alternate_screen;controlled_alt_requested=[bool]$Alternate;scroll_offset=$p.scroll_offset;view_offset=$p.view_offset;copy_cursor_row=$p.copy_cursor_row;copy_cursor_col=$p.copy_cursor_col;sel_start_row=$p.sel_start_row;sel_end_row=$p.sel_end_row;sel_mode=$p.sel_mode;screen_text=@($p.rows_v2|ForEach-Object{($_.runs.text -join '').TrimEnd()});status=$f.status_message}
 if($Alternate){$started=Join-Path $testRoot 'work/alt-started.json';if(!(Test-Path $started) -or (($o.screen_text -join "`n") -notmatch 'VT_ALT_READY')){throw 'Controlled VT application not independently observed'};$id=[IO.File]::ReadAllText($started)|ConvertFrom-Json;$child=Get-Process -Id $id.pid;if($child.Path -ne $id.path -or $child.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime){throw 'Controlled VT process changed'}}
 if(!$Alternate){
  $data=Join-Path (Join-Path $testRoot 'runtime') $ns;$env:PSMUX_DATA_DIR=$data;$env:PSMUX_CONFIG_FILE=Join-Path $data 'psmux-m0.conf';$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1'
  $text=(& $backend -L $ns capture-pane -t ($s.backend_id+':.%'+$p.id) -p -S -200|Out-String)
  if($LASTEXITCODE -ne 0){throw 'Native capture failed'}
  $ids=@(($text -split '\r?\n')|ForEach-Object{if($_ -match '^history line (\d+)$'){[int]$Matches[1]}})
  $o.capture_text=$text;$o.numbered_ids=$ids;$o.complete=(($ids -join ',') -ceq ((1..40) -join ','))
  $script:observations.Add($o);$script:observations.ToArray()|ConvertTo-Json -Depth 10|Set-Content (Join-Path $testRoot 'native-observations.json') -Encoding utf8
  if(!$o.complete){throw 'Retained numbered output is missing or duplicated'}
 }else{$script:observations.Add($o);$script:observations.ToArray()|ConvertTo-Json -Depth 10|Set-Content (Join-Path $testRoot 'native-observations.json') -Encoding utf8}
 return $o
}
function CheckShells{
 $ids=@();foreach($name in @('Copy_A','Copy_B')){$i=[IO.File]::ReadAllText((Join-Path $testRoot ('work/'+$name+'.identity.json')))|ConvertFrom-Json;$p=Get-Process -Id $i.pid;if($p.Path -ne $i.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $i.filetime){throw 'Original shell changed'};$ids+=,$i};return $ids
}
function CheckDraft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title
 if($o.capture_text -notmatch [regex]::Escape($draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw 'Original draft was changed or submitted before explicit Enter'}
}

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

function WaitRootOrder([string]$first,[string]$second){
 $until=[DateTime]::UtcNow.AddSeconds(10)
 do{$text=StateText;$a=$text.IndexOf($first+' (0)');$b=$text.IndexOf($second+' (0)');if($a -ge 0 -and $b -ge 0 -and $a -lt $b){return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until)
 throw 'Visible root order not observed'
}
function GroupById($m,[string]$id){$a=@($m.Groups|Where-Object{$_.ID -ceq $id});if($a.Count -ne 1){throw 'Group identity not unique'};return $a[0]}
function Roots($m){return ((@($m.Groups|Where-Object{$_.ParentID -eq ''}|Sort-Object Order)).ID -join ',')}
function OrderMeta([string]$label,[string]$roots,[int]$rootRev,[int]$childRev=1){
 $m=Meta $label
 if(@($m.Groups).Count -ne 5 -or (Roots $m) -cne $roots -or ($m.Members|ConvertTo-Json -Depth 8 -Compress) -cne $initialMembers -or ($m.Runs|ConvertTo-Json -Depth 8 -Compress) -cne $initialRuns){throw ('Order changed tree/member/Run contract: '+$label)}
 foreach($original in @($rootA,$rootB,$review,$implementation)){
  $now=GroupById $m $original.ID
  if($now.Name -cne $original.Name -or $now.ParentID -cne $original.ParentID -or $now.DefaultDir -cne $original.DefaultDir){throw ('Sort changed group identity/config: '+$label)}
  $rev=$rootRev;if($original.ParentID){$rev=$childRev};if($now.Revision -ne $rev){throw ('Unexpected group revision: '+$label)}
 }
 $def=GroupById $m 'default';if($def.Order -ne 0 -or $def.Revision -ne 1){throw 'Default group moved or changed'}
 return $m
}
function WaitChildOrder{
 $until=[DateTime]::UtcNow.AddSeconds(10)
 do{$text=StateText;if($text.IndexOf('实现 (0)') -ge 0 -and $text.IndexOf('实现 (0)') -lt $text.IndexOf('审核 (1)')){return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until)
 throw 'Child order not observed'
}
$a="Write-Output ('ORDER_A:'+`$PID) # KJ+-";$b="Write-Output ('ORDER_B:'+`$PID)";$originalSessions=@($sessions)
try{
 $null=WaitText 'COPY_READY Copy A';RememberUI;$before=@(CheckShells)
 Paste "Write-Output ('ORDER_A:'+`$PID) # ";Text 'KJ+-';Draft 'A-initial-terminal-literals' 'Copy A' $a 'ORDER_A'
 Key '^{PGDN}';$null=WaitText 'COPY_READY Copy B';Paste $b;Draft 'B-initial' 'Copy B' $b 'ORDER_B';Key '^{PGUP}';$null=WaitText ([regex]::Escape($a));Key '{F6}';$null=WaitText '管理模式'
 Text 'g';$null=WaitText '创建分组';Text '项目 甲';Key '{ENTER}';$null=WaitText '项目 甲 \(0\)'
 $m=Meta 'root-A-created';$rootA=@($m.Groups|Where-Object{$_.Name -ceq '项目 甲'})[0];if(!$rootA){throw 'Root A missing'}
 Text 'g';$null=WaitText '创建子组';Text '审核';Key '{ENTER}';$null=WaitText '审核 \(0\)'
 $m=Meta 'review-created';$review=@($m.Groups|Where-Object{$_.Name -ceq '审核'})[0];if(!$review){throw 'Review missing'}
 Text 'g';$null=WaitText '创建子组';Key '{UP}';$null=WaitText '创建分组';Text '项目 乙';Key '{ENTER}';$null=WaitText '项目 乙 \(0\)'
 $m=Meta 'root-B-created';$rootB=@($m.Groups|Where-Object{$_.Name -ceq '项目 乙'})[0];if(!$rootB){throw 'Root B missing'}
 Key '2';Text 'g';$null=WaitText '创建子组';Text '实现';Key '{ENTER}';$null=WaitText '实现 \(0\)'
 $m=Meta 'implementation-created';$implementation=@($m.Groups|Where-Object{$_.Name -ceq '实现'})[0];if(!$implementation -or $implementation.ParentID -cne $rootA.ID){throw 'Implementation parent differs'}
 Key '{HOME}';Key '{DOWN}';Text 'M';$null=WaitText '移动到分组';Key '{DOWN}';Key '{DOWN}';Key '{ENTER}';$null=WaitText '审核 \(1\)'
 $initial=Meta 'before-order';$initialMembers=$initial.Members|ConvertTo-Json -Depth 8 -Compress;$initialRuns=$initial.Runs|ConvertTo-Json -Depth 8 -Compress
 $rootsAB='default,'+$rootA.ID+','+$rootB.ID;$rootsBA='default,'+$rootB.ID+','+$rootA.ID
 $null=OrderMeta 'initial-contract' $rootsAB 1;$null=Shot 'order-initial-subtree-120-public'
 Key '3';Text 'K';WaitRootOrder '项目 乙' '项目 甲';$null=OrderMeta 'K-root-up' $rootsBA 2;$null=Shot 'order-root-K-up-120-public'
 Key '{LEFT}';$null=WaitText '▸ 项目 乙';Pref 'B-folded' 'compact' $rootB.ID;$foldHash=(Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash
 Text 'J';WaitRootOrder '项目 甲' '项目 乙';$null=OrderMeta 'J-root-down' $rootsAB 3
 Key '^{UP}';WaitRootOrder '项目 乙' '项目 甲';$null=OrderMeta 'ctrl-up-root' $rootsBA 4
 Key '^{DOWN}';WaitRootOrder '项目 甲' '项目 乙';$null=OrderMeta 'ctrl-down-root' $rootsAB 5
 Key '+{UP}';WaitRootOrder '项目 乙' '项目 甲';$null=OrderMeta 'shift-up-root' $rootsBA 6
 Key '+{DOWN}';WaitRootOrder '项目 甲' '项目 乙';$null=OrderMeta 'shift-down-root' $rootsAB 7
 Text '+';WaitRootOrder '项目 乙' '项目 甲';$null=OrderMeta 'plus-root-up' $rootsBA 8
 Text '-';WaitRootOrder '项目 甲' '项目 乙';$last=OrderMeta 'minus-root-down' $rootsAB 9;$null=Shot 'order-eight-aliases-120-public'
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $foldHash){throw 'Sort changed folded preferences'}
 Draft 'A-after-root-order' 'Copy A' $a 'ORDER_A';Draft 'B-after-root-order' 'Copy B' $b 'ORDER_B';$null=@(CurrentSessions);$null=@(CheckShells)
 Key '1';Text 'K';Text 'J';Key '2';Text 'K';Key '3';Text 'J'
 $bounded=OrderMeta 'pinned-and-boundary-noop' $rootsAB 9
 if(($bounded.Groups|ConvertTo-Json -Depth 8 -Compress) -cne ($last.Groups|ConvertTo-Json -Depth 8 -Compress)){throw 'Boundary/default edits changed versions'}
 Key '2';Key '{DOWN}';Key '{DOWN}';Key '{DOWN}';Text 'K';WaitChildOrder
 $m=OrderMeta 'child-K-up' $rootsAB 9 2
 $i=GroupById $m $implementation.ID;$v=GroupById $m $review.ID;if($i.Order -ne 0 -or $v.Order -ne 1){throw 'Child swap crossed parent or failed'}
 $null=Shot 'order-child-up-120-public'
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '管理模式';ViewCols 63;$null=Shot 'order-grouped-63-public'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '管理模式';ViewCols 49;$null=Shot 'order-grouped-49-public'
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText '管理模式';ViewCols 120
 Key '2';Key '{LEFT}';$null=WaitText '▸ 项目 甲';$folds=(@($rootA.ID,$rootB.ID)|Sort-Object)-join ',';Pref 'both-folded-before-reopen' 'compact' $folds;$foldHash=(Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash
 $backup=Join-Path $testRoot 'order-backup.sqlite3';$null=& $product @g metadata backup $backup;if($LASTEXITCODE -ne 0){throw 'Metadata backup failed'}
 $backupMeta=Meta 'backup-order' $backup
 if(($backupMeta.Groups|ConvertTo-Json -Depth 8 -Compress) -cne ($m.Groups|ConvertTo-Json -Depth 8 -Compress)){throw 'Backup omitted sorted groups'}
 Text 'q';WaitExit;$null=@(CurrentSessions);Draft 'A-after-quit' 'Copy A' $a 'ORDER_A';Draft 'B-after-quit' 'Copy B' $b 'ORDER_B'
 $launch="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $launch;Key '{ENTER}';$null=WaitText '终端模式';RememberUI;Key '{F6}';$null=WaitText '管理模式';$null=WaitText '▸ 项目 甲';WaitRootOrder '项目 甲' '项目 乙'
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $foldHash){throw 'Reopen changed folds'}
 $m=OrderMeta 'reopened-order' $rootsAB 9 2;$null=Shot 'order-folded-reopen-120-public'
 Text '/';$null=WaitText '本地会话搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText '终端模式';Draft 'A-before-explicit-enter' 'Copy A' $a 'ORDER_A';Key '{ENTER}';$null=WaitText ('ORDER_A:'+$before[0].pid);$oa=Observe 'A-explicit-enter' 'Copy A';if(@([regex]::Matches($oa.capture_text,('(?m)^ORDER_A:'+$before[0].pid+'\s*$'))).Count -ne 1){throw 'A executed elsewhere/more than once'}
 Key '^{PGDN}';$null=WaitText '终端模式';Draft 'B-before-explicit-enter' 'Copy B' $b 'ORDER_B';Key '{ENTER}';$null=WaitText ('ORDER_B:'+$before[1].pid);$ob=Observe 'B-explicit-enter' 'Copy B';if(@([regex]::Matches($ob.capture_text,('(?m)^ORDER_B:'+$before[1].pid+'\s*$'))).Count -ne 1){throw 'B executed elsewhere/more than once'}
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit;$after=@(CurrentSessions);if($after.Count -ne 2){throw 'UI exit changed live set'};$null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;before=$originalSessions;after=$after;before_processes=$before;after_processes=@(CheckShells);ui_processes=$script:uis.ToArray();root_A=$rootA;root_B=$rootB;review=$review;implementation=$implementation;final_groups=$m.Groups;root_aliases=@('K','J','ctrl+up','ctrl+down','shift+up','shift+down','+','-');root_swaps=8;child_swaps=1;preference_snapshots=$script:preferences.Count;metadata_snapshots=$script:metadata.Count;scope='Two ordinary shells; root aliases, expanded-subtree skipping, child swap, pinned/boundary no-op, 120/63/49 columns, snapshot/reopen, original drafts/PIDs. Unicode literals are not physical uppercase or IME-composition acceptance; async fault barriers tested in UI models separately.'}|ConvertTo-Json -Depth 12|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8;throw}
