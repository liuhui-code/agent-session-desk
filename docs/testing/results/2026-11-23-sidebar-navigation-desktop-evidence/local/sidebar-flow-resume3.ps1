$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'sidebar-navigation-desktop-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'asd-sidebar-navigation.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-sidebar-navigation';$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
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
$script:preferences=[Collections.Generic.List[object]]::new()
function Pref([string]$label,[string]$density,[string]$mode,[bool]$collapsed){
 $path=Join-Path $testRoot 'state/ui.toml';$until=[DateTime]::UtcNow.AddSeconds(8)
 do{$text=ReadSharedText $path;$ok=$text -match '(?m)^schema_version = 2$' -and $text -match ('(?m)^sidebar_density = "'+$density+'"$') -and $text -match ('(?m)^sidebar_mode = "'+$mode+'"$') -and $text -match ('(?m)^default_group_collapsed = '+$collapsed.ToString().ToLowerInvariant()+'$');if($ok){$script:preferences.Add(@{label=$label;density=$density;mode=$mode;collapsed=$collapsed;text=$text;sha256=(Get-FileHash $path).Hash});$script:preferences.ToArray()|ConvertTo-Json -Depth 6|Set-Content (Join-Path $testRoot 'preferences.json') -Encoding utf8;return};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
 throw ('Preference not observed: '+$label)
}
function ClickNav([string]$title,[bool]$metadata=$false){$null=InvokeVisual @{kind='clickNavRow';title=$title;metadata=$metadata}}
function Draft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title;$lines=@(($o.capture_text -split '\r?\n')|Where-Object{$_.Trim()})
 if(!$lines.Count -or $lines[-1].Trim() -cne ('> '+$draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw ('Pending original draft differs: '+$label)}
}
function ViewCols([int]$cols){
 $text=StateText;$lines=@(($text -split '\r?\n')|Where-Object{$_ -match 'Copy A ×' -and $_ -match 'Copy B ×'});if($lines.Count -ne 1){throw 'Unique application Tab row not observed'};$line=$lines[0]
 if($line.Length -ne $cols){throw ('Actual terminal columns differ: '+$line.Length+' versus '+$cols)}
}
$a="Write-Output ('NAV_A:'+`$PID)";$b="Write-Output ('NAV_B:'+`$PID)"
$script:sequence=152
foreach($o in ([IO.File]::ReadAllText((Join-Path $testRoot 'native-observations.json'))|ConvertFrom-Json)){$script:observations.Add($o);$script:panes[$o.session_id]=$o.pane}
foreach($id in ([IO.File]::ReadAllText((Join-Path $testRoot 'ui-identities.json'))|ConvertFrom-Json)){$script:uis.Add($id)}
foreach($pref in ([IO.File]::ReadAllText((Join-Path $testRoot 'preferences.json'))|ConvertFrom-Json)){$script:preferences.Add($pref)}
$before=@(CheckShells);$seed=(Get-FileHash (Join-Path $testRoot 'ui-v1-before.toml')).Hash
[IO.File]::Copy((Join-Path $testRoot 'flow-failure.json'),(Join-Path $testRoot 'flow-failure-active-marker.json'))
try{
 $null=WaitText '[▶┃] Copy B';Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText ([regex]::Escape($b));Draft 'B-wheel-open' 'Copy B' $b 'NAV_B';Key '{F6}';$null=WaitText '管理模式'
$null=Shot 'grouped-wheel-selection-public'
 Text 'S';$null=WaitText '侧栏密度';Key '{RIGHT}';Key '{ESC}';$null=WaitText '管理模式';Pref 'density-auto' 'auto' 'grouped' $false
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '▾ 我的会话';ViewCols 63;$null=Shot 'grouped-auto-63-public';ClickNav 'Copy A' $true;$null=WaitText '▶ Copy A'
 $null=InvokeVisual @{kind='resize';width=400;height=539};$null=WaitText '▾ 我的会话';ViewCols 49;$null=Shot 'grouped-auto-49-public'
 ClickNav '我的会话';$null=WaitText '▸ 我的会话';Pref 'narrow-mouse-fold' 'auto' 'grouped' $true
 ClickNav '我的会话';$null=WaitText '▾ 我的会话';Pref 'narrow-mouse-expand' 'auto' 'grouped' $false
 ClickNav 'Copy B' $true;$null=WaitText '[▶┃] Copy B';Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText ([regex]::Escape($b));Draft 'B-narrow-open' 'Copy B' $b 'NAV_B';$null=Shot 'terminal-49-original-draft-public'
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText '终端模式';Key '{F6}';$null=WaitText '管理模式';Key '1';Key '{ENTER}';$null=WaitText '▸ 我的会话';Pref 'before-search-fold' 'auto' 'grouped' $true
 $foldHash=(Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash
 Text '/';$null=WaitText '本地会话搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText '▾ 我的会话';Draft 'A-search-reveal' 'Copy A' $a 'NAV_A'
 if((Get-FileHash (Join-Path $testRoot 'state/ui.toml')).Hash -ne $foldHash){throw 'Search reveal unexpectedly saved expansion'}
 Key '{F6}';$null=WaitText '管理模式';Text 'S';$null=WaitText '侧栏密度';Key '{LEFT}';Key '{LEFT}';Key '{LEFT}';Key '{ESC}';$null=WaitText '管理模式';Pref 'final-density-full' 'full' 'grouped' $false
 Key '1';Key '{ENTER}';Text 'q';WaitExit;Pref 'quit-saved-fold' 'full' 'grouped' $true
 $middle=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $middle.Count -ne 2){throw 'UI quit changed live Session set'}
 foreach($s in $sessions){if(@($middle|Where-Object{$_.session_id -eq $s.session_id -and $_.run_id -eq $s.run_id -and $_.backend_id -eq $s.backend_id}).Count -ne 1){throw 'UI quit changed original Run'}}
 Draft 'A-after-UI-quit' 'Copy A' $a 'NAV_A';Draft 'B-after-UI-quit' 'Copy B' $b 'NAV_B'
 $launch="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $launch;Key '{ENTER}';$null=WaitText '终端模式';RememberUI;Key '{F6}';$null=WaitText '管理模式';$null=WaitText '▸ 我的会话';Pref 'UI-reopen-retains-fold-density-mode' 'full' 'grouped' $true;$null=Shot 'grouped-full-folded-reopen-public'
 Key '{ENTER}';$null=WaitText '▾ 我的会话';Key '{DOWN}';$null=WaitText '▶ Copy A';Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText ([regex]::Escape($a));Draft 'A-reopen-before-enter' 'Copy A' $a 'NAV_A'
 Key '{ENTER}';$null=WaitText ('NAV_A:'+$before[0].pid);$oa=Observe 'A-explicit-enter' 'Copy A';if($oa.capture_text -notmatch ('(?m)^NAV_A:'+$before[0].pid+'\s*$')){throw 'A command ran in another process'}
 Key '^{PGDN}';$null=WaitText '终端模式';$null=WaitText ([regex]::Escape($b));Draft 'B-reopen-before-enter' 'Copy B' $b 'NAV_B';Key '{ENTER}';$null=WaitText ('NAV_B:'+$before[1].pid);$ob=Observe 'B-explicit-enter' 'Copy B';if($ob.capture_text -notmatch ('(?m)^NAV_B:'+$before[1].pid+'\s*$')){throw 'B command ran in another process'}
 Key '{F6}';$null=WaitText '管理模式';Key '%v';$null=WaitText '会话 平铺';Pref 'final-flat' 'full' 'flat' $false;$null=Shot 'flat-full-final-public'
 Text 'q';WaitExit;$after=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $after.Count -ne 2){throw 'Final UI quit changed Session set'}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;legacy_read_hash=$seed;before=$sessions;after=$after;before_processes=$before;after_processes=@(CheckShells);ui_processes=$script:uis.ToArray();preference_snapshots=$script:preferences.Count;scope='Two ordinary shells; current grouped navigation, 120/63/49 columns, density modes, native mouse/wheel and original drafts across UI reopen. No real agent or full release acceptance.'}|ConvertTo-Json -Depth 10|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8;throw}
