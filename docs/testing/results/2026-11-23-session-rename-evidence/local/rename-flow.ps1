$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'session-rename-desktop-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'asd-rename.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-session-rename';$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
$sessions=([IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json);$script:panes=@{};$script:sequence=0;$script:observations=[Collections.Generic.List[object]]::new();$script:uis=[Collections.Generic.List[object]]::new()
function Action($id,$value){
 $name=('{0:d3}.json' -f $id);$path=Join-Path $root "commands/$name";$temp=Join-Path $root "$name.tmp"
 [IO.File]::WriteAllText($temp,($value|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false));Move-Item $temp $path
 $response=Join-Path $root "responses/$name";$until=[DateTime]::UtcNow.AddSeconds(15)
 while(!(Test-Path $response) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $response)){throw 'Driver response timeout'};$r=[IO.File]::ReadAllText($response)|ConvertFrom-Json
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
 $port=[int]([IO.File]::ReadAllText((Join-Path $data ($reg+'.port')))).Trim();$key=([IO.File]::ReadAllText((Join-Path $data ($reg+'.key')))).Trim()
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
 $ids=@();foreach($name in @('Copy_A','Copy_B','中文会话')){$i=[IO.File]::ReadAllText((Join-Path $testRoot ('work/'+$name+'.identity.json')))|ConvertFrom-Json;$p=Get-Process -Id $i.pid;if($p.Path -ne $i.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $i.filetime){throw 'Original shell changed'};$ids+=,$i};return $ids
}
function CheckDraft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title
 if($o.capture_text -notmatch [regex]::Escape($draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw 'Original draft was changed or submitted before explicit Enter'}
}
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
