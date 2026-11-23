$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'main-allmotion-desktop-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'asd-main-allmotion.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-main-allmotion';$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
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
$originalSessions=@($sessions)
function Draft([string]$label,[string]$title,[string]$draft,[string]$output){$o=Observe $label $title;$lines=@(($o.capture_text -split '\r?\n')|Where-Object{$_.Trim()});if(!$lines.Count -or $lines[-1].Trim() -cne ('> '+$draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw ('Original draft differs: '+$label)}}
function CurrentSessions{$rows=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or @($rows).Count -ne 2){throw 'Live set changed'};foreach($s in $originalSessions){if(@($rows|Where-Object{$_.session_id -ceq $s.session_id -and $_.run_id -ceq $s.run_id -and $_.backend_id -ceq $s.backend_id}).Count -ne 1){throw 'Original identity changed'}};return $rows}
function ViewCols([int]$cols){$text=StateText;$lines=@(($text -split '\r?\n')|Where-Object{$_ -match 'Copy A ×' -and $_ -match 'Copy B ×'});if($lines.Count -ne 1 -or $lines[0].Length -ne $cols){throw 'Viewport width differs'}}
function NarrowFooter{$text=WaitText '终端模式';$line=@(($text -split '\r?\n')|Where-Object{$_ -match '终端模式'})[-1];if($line -match '→|\]' -or $line -notmatch 'F6' -or $line -notmatch 'Ctrl\+PgUp/PgDn'){throw ('Narrow target fragment/navigation: '+$line)}}

function ClickNav([string]$title,[bool]$metadata=$false){$null=InvokeVisual @{kind='clickNavRow';title=$title;metadata=$metadata}}
function Divider([string]$phase){$null=InvokeVisual @{kind='divider';phase=$phase}}
function RowSize([int]$cols,[int]$rows){
 $text=StateText;$lines=$text -split '\r?\n';$start=-1
 for($i=0;$i -lt $lines.Length;$i++){if($lines[$i].StartsWith('Agent Session Desk')){$start=$i;break}}
 if($start -lt 0){throw 'App header not observed'}
 $tab=$lines[$start+1];$actual=$tab.Length+[regex]::Matches($tab,'[\u4e00-\u9fff]').Count
 if($actual -ne $cols){throw ('Actual columns differ: '+$actual)}
 $body=@($lines[($start+2)..($lines.Length-1)]);$rule=-1
 for($i=0;$i -lt $body.Count;$i++){if($body[$i] -ceq ('─'*$cols)){$rule=$i;break}}
 if($rule -ne $rows-5){throw ('Actual row band differs: '+$rule)}
}
$a="Write-Output ('MAIN_A:'+`$PID) # native";$b="Write-Output ('MAIN_B:'+`$PID)"
$pref=Join-Path $testRoot 'state/ui.toml';$prefHash=(Get-FileHash $pref).Hash
$task=Get-ScheduledTask 'AsdMainAllMotionDesktop20261123';if($task.State -ne 'Running'){throw 'Original task no longer running'}
$drivers=@();foreach($q in @(Get-CimInstance Win32_Process|Where-Object{$_.Name -eq 'powershell.exe' -and $_.SessionId -eq 2})){
 if($q.CommandLine -match '(?i)-EncodedCommand\s+([A-Za-z0-9+/=]+)'){$decoded=[Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($Matches[1]));if($decoded.Contains('main-allmotion-desktop-driver.ps1')){$d=Get-Process -Id $q.ProcessId;$drivers+=,@{pid=$d.Id;filetime=$d.StartTime.ToUniversalTime().ToFileTimeUtc();path=$d.Path}}}
}
if($drivers.Count -ne 1){throw 'Live original driver not uniquely observed'}
$last=Get-ChildItem (Join-Path $root 'responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1
if($last.Name -cne '074.json' -or @(Get-ChildItem (Join-Path $root 'commands') -File).Count){throw 'Original sequence changed'}
$ready=[IO.File]::ReadAllText((Join-Path $root 'ready.json'))|ConvertFrom-Json;$win=Get-Process -Id $ready.pid
if($win.Path -ne $ready.path -or $win.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $ready.filetime){throw 'Original window changed'}
$u=[IO.File]::ReadAllText((Join-Path $testRoot 'ui-identities.json'))|ConvertFrom-Json
foreach($id in $u){$p=Get-Process -Id $id.pid;if($p.Path -ne $id.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime){throw 'Original UI changed'};$script:uis.Add($id)}
$old=[IO.File]::ReadAllText((Join-Path $testRoot 'native-observations.json'))|ConvertFrom-Json
foreach($o in $old){$script:observations.Add($o);$script:panes[$o.session_id]=$o.pane}
if($script:observations.Count -ne 2){throw 'Original captures differ'}
$before=@(CheckShells);$script:sequence=74
try{
 $fresh=StateText;if($fresh -notmatch '终端模式' -or $fresh -notmatch '→ Copy B' -or $fresh -notmatch [regex]::Escape($b)){throw 'Fresh state not original terminal B'}
 @{utc=[DateTime]::UtcNow.ToString('o');original_driver=$drivers[0];previous_sequence=74;fresh_sequence=$script:sequence;replayed_mutations=0;reason='Ctrl+PgDown preserved management focus; the accepted F6 entered terminal B. Continue from observed terminal B with one new explicit F6 to management, without repeating original switch or drafts.'}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-resume.json') -Encoding utf8
 Key '{F6}';$null=WaitText '管理模式';
ClickNav 'Copy A' $true;RowSize 120 32
 Divider 'leave';$null=Shot 'main-nested-selected-120-public';Draft 'A-nested-selected' 'Copy A' $a 'MAIN_A';Draft 'B-management-active' 'Copy B' $b 'MAIN_B'
 Divider 'hover';$null=WaitText '会话 分组.* ┃ ';$null=Shot 'main-divider-hover-120-public'
 Divider 'press';$null=Shot 'main-divider-pressed-120-public';Divider 'release';Divider 'leave';$null=WaitText '会话 分组.* │ '
 Key '{ENTER}';$null=WaitText 'COPY_READY Copy A';$null=WaitText '终端模式';$null=Shot 'main-nested-active-120-public';Draft 'A-nested-active' 'Copy A' $a 'MAIN_A'
 $null=InvokeVisual @{kind='resize';width=1473;height=657};RowSize 160 32;$null=Shot 'main-nested-active-160-public'
 Key '{F6}';$null=WaitText '管理模式';$null=Shot 'main-nested-selected-160-public'
 $null=InvokeVisual @{kind='resize';width=600;height=539};RowSize 63 25;$null=Shot 'main-nested-manager-63-public';Draft 'A-stacked' 'Copy A' $a 'MAIN_A'
 $null=InvokeVisual @{kind='resize';width=400;height=539};RowSize 49 25;$null=Shot 'main-nested-manager-49-public';Draft 'A-single' 'Copy A' $a 'MAIN_A'
 $null=InvokeVisual @{kind='resize';width=1113;height=277};RowSize 120 12;$null=Shot 'main-nested-manager-low-120-public';Draft 'A-low-height' 'Copy A' $a 'MAIN_A'
 $null=InvokeVisual @{kind='resize';width=1113;height=657};RowSize 120 32;Key '{ENTER}';$null=WaitText 'COPY_READY Copy A';Draft 'A-restored' 'Copy A' $a 'MAIN_A'
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit;$null=@(CurrentSessions)
 $launch="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Text $launch;Key '{ENTER}';$null=WaitText '终端模式';RememberUI;Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '本地会话搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText 'COPY_READY Copy A';Key '{F6}';$null=WaitText '管理模式';$null=WaitText '▾ 审核 \(1\)';$null=Shot 'main-nested-reopened-120-public';Draft 'A-reopened' 'Copy A' $a 'MAIN_A';Draft 'B-reopened' 'Copy B' $b 'MAIN_B'
 Key '{ENTER}';$null=WaitText '终端模式';Key '{ENTER}';$null=WaitText ('MAIN_A:'+$before[0].pid);$oa=Observe 'A-explicit-enter' 'Copy A';if(@([regex]::Matches($oa.capture_text,('(?m)^MAIN_A:'+$before[0].pid+'\s*$'))).Count -ne 1){throw 'A executed elsewhere/more than once'}
 Key '^{PGDN}';$null=WaitText 'COPY_READY Copy B';Draft 'B-before-enter' 'Copy B' $b 'MAIN_B';Key '{ENTER}';$null=WaitText ('MAIN_B:'+$before[1].pid);$ob=Observe 'B-explicit-enter' 'Copy B';if(@([regex]::Matches($ob.capture_text,('(?m)^MAIN_B:'+$before[1].pid+'\s*$'))).Count -ne 1){throw 'B executed elsewhere/more than once'}
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit;$after=@(CurrentSessions);if((Get-FileHash $pref).Hash -cne $prefHash){throw 'Rendering or group creation changed UI preferences'};$null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;before=$originalSessions;after=$after;before_processes=$before;after_processes=@(CheckShells);ui_processes=$script:uis.ToArray();native_captures=$script:observations.Count;preferences_before=$prefHash;preferences_after=(Get-FileHash $pref).Hash;scope='Actual two-shell nested creation/move, selected versus active, 120/160/63/49 columns and 12 rows, divider hover/press/release, original drafts and 40 retained output lines across UI restart. No real agent or complete visual acceptance.'}|ConvertTo-Json -Depth 12|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'resume-failure.json') -Encoding utf8;throw}
