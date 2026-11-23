$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'session-search-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'search-product.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-session-search';$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
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
