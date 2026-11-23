$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'footer-input-desktop-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'asd-footer-input.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-footer-input';$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
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
$originalSessions=@($sessions);$script:preferences=[Collections.Generic.List[object]]::new()
function Draft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title;$lines=@(($o.capture_text -split '\r?\n')|Where-Object{$_.Trim()})
 if(!$lines.Count -or $lines[-1].Trim() -cne ('> '+$draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw ('Original pending draft differs: '+$label)}
}
function CurrentSessions{$rows=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or @($rows).Count -ne 2){throw 'Live set changed'};foreach($s in $originalSessions){if(@($rows|Where-Object{$_.session_id -ceq $s.session_id -and $_.run_id -ceq $s.run_id -and $_.backend_id -ceq $s.backend_id -and $_.command -ceq $s.command}).Count -ne 1){throw 'Original identity changed'}};return $rows}
function ViewCols([int]$cols){$text=StateText;$lines=@(($text -split '\r?\n')|Where-Object{$_ -match 'Copy A ×' -and $_ -match 'Copy B ×'});if($lines.Count -ne 1 -or $lines[0].Length -ne $cols){throw ('Unique Tab row width differs from '+$cols)}}
function SharedHash($path){$stream=[IO.File]::Open($path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete));try{$hash=[Security.Cryptography.SHA256]::Create();try{return ([BitConverter]::ToString($hash.ComputeHash($stream))).Replace('-','').ToLowerInvariant()}finally{$hash.Dispose()}}finally{$stream.Dispose()}}
$prefPath=Join-Path $testRoot 'state/ui.toml';$prefHash=SharedHash $prefPath
function Pref([string]$label){$until=[DateTime]::UtcNow.AddSeconds(10);do{$hash=SharedHash $prefPath;if($hash -ceq $prefHash){$script:preferences.Add(@{label=$label;sha256=$hash});$script:preferences.ToArray()|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'preference-observations.json') -Encoding utf8;return};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until);throw ('Preference changed unexpectedly: '+$label)}
$a="Write-Output ('FOOT_A:'+`$PID) # nrMdx?+";$b="Write-Output ('FOOT_B:'+`$PID)"

$task=Get-ScheduledTask 'AsdFooterInputDesktop20261123'
$driver=Get-Process -Id 10004 -ErrorAction Stop
if($task.State -ne 'Running' -or $driver.StartTime.ToUniversalTime().ToFileTimeUtc() -ne 134398946605874415 -or $driver.Path -cne 'C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe'){throw 'Original driver not live; no input'}
$last=Get-ChildItem (Join-Path $root 'responses') -Filter '*.json'|Sort-Object Name|Select-Object -Last 1
if($last.Name -cne '154.json'){throw 'Unexpected completed sequence; no input'}
$lastResult=ReadResponse $last.FullName;if(!$lastResult.ok -or $lastResult.command.kind -ne 'state'){throw 'Last observation not complete'}
$oldUI=@((ReadSharedText (Join-Path $testRoot 'ui-identities.json'))|ConvertFrom-Json)
if($oldUI.Count -ne 1){throw 'Expected exactly one original UI'}
$ui=Get-Process -Id $oldUI[0].pid;if($ui.Path -ne $oldUI[0].path -or $ui.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $oldUI[0].filetime){throw 'Original UI changed'}
$script:uis.Add($oldUI[0]);$before=@(CheckShells)
$obs=@((ReadSharedText (Join-Path $testRoot 'native-observations.json'))|ConvertFrom-Json)
if($obs.Count -ne 4 -or ($obs.label -join ',') -cne 'A-initial-terminal-letters,B-initial,A-after-help-pastes,B-after-help-pastes'){throw 'Previous original captures incomplete'}
foreach($o in $obs){$script:observations.Add($o);if($script:panes.ContainsKey($o.session_id) -and $script:panes[$o.session_id] -ne $o.pane){throw 'Original pane changed'};$script:panes[$o.session_id]=$o.pane}
$preferences=@((ReadSharedText (Join-Path $testRoot 'preference-observations.json'))|ConvertFrom-Json)
if($preferences.Count -ne 12 -or @($preferences|Where-Object{$_.label -like 'paste-*'}).Count -ne 10){throw 'Previous paste cases incomplete'}
foreach($v in $preferences){$script:preferences.Add($v)}
$prefHash=SharedHash (Join-Path $testRoot 'ui-v3-before.toml');if((SharedHash $prefPath) -cne $prefHash){throw 'Preferences changed'}
$case=10;$script:sequence=154
try{
 $fresh=StateText;$rows=@(($fresh -split '\r?\n')|Where-Object{$_ -match 'Copy B ×'})
 if($rows.Count -ne 1 -or $fresh -match 'Copy A ×' -or $fresh -notmatch '管理模式' -or $fresh -notmatch [regex]::Escape($b)){throw 'Fresh one-original-B-Tab observation failed; no closing input'}
 $null=@(CurrentSessions)
 @{previous_completed_sequence=154;fresh_state_sequence=$script:sequence;original_driver_pid=$driver.Id;original_driver_filetime=$driver.StartTime.ToUniversalTime().ToFileTimeUtc();replayed_mutations=0;accepted_first_close_original=$originalSessions[0];remaining_tab='Copy B';reason='First x already closed A; connection completion cleared its transient status before the old wait. Continue with the intended second close only.'}|ConvertTo-Json -Depth 9|Set-Content (Join-Path $testRoot 'flow-resume.json') -Encoding utf8
 $null=Shot 'footer-one-tab-after-accepted-close-120-public'
 Text 'x';$null=WaitText '暂无打开的 Tab';
Key '{F6}';$null=WaitText '管理模式';$null=Shot 'footer-no-open-tabs-120-public';$null=@(CurrentSessions)
 Text '/';$null=WaitText '搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText '→ Copy A';$null=Shot 'footer-reopened-tab-A-120-public';Draft 'A-after-tab-reopen' 'Copy A' $a 'FOOT_A';Draft 'B-after-tab-reopen' 'Copy B' $b 'FOOT_B';Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit;$null=@(CurrentSessions)
 Draft 'A-after-UI-quit' 'Copy A' $a 'FOOT_A';Draft 'B-after-UI-quit' 'Copy B' $b 'FOOT_B'
 $launch="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $launch;Key '{ENTER}';$null=WaitText '终端模式';RememberUI;Key '{F6}';$null=WaitText '管理模式';Text '/';$null=WaitText '搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText '→ Copy A';$null=Shot 'footer-reopened-UI-A-120-public';Pref 'after-UI-reopen'
 Draft 'A-before-explicit-enter' 'Copy A' $a 'FOOT_A';Key '{ENTER}';$null=WaitText ('FOOT_A:'+$before[0].pid);$oa=Observe 'A-explicit-enter' 'Copy A';if(@([regex]::Matches($oa.capture_text,('(?m)^FOOT_A:'+$before[0].pid+'\s*$'))).Count -ne 1){throw 'A executed elsewhere/more than once'}
 Key '^{PGDN}';$null=WaitText '→ Copy B';Draft 'B-before-explicit-enter' 'Copy B' $b 'FOOT_B';Key '{ENTER}';$null=WaitText ('FOOT_B:'+$before[1].pid);$ob=Observe 'B-explicit-enter' 'Copy B';if(@([regex]::Matches($ob.capture_text,('(?m)^FOOT_B:'+$before[1].pid+'\s*$'))).Count -ne 1){throw 'B executed elsewhere/more than once'}
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit;$after=@(CurrentSessions);$null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;before=$originalSessions;after=$after;before_processes=$before;after_processes=@(CheckShells);ui_processes=$script:uis.ToArray();native_captures=$script:observations.Count;help_paste_cases=$case;preference_snapshots=$script:preferences.Count;explicit_fold_toggles=2;preferences_before=$prefHash;preferences_after=(SharedHash $prefPath);scope='Two original PowerShells; management footer/group fold contexts, 120/160/80/100/101/63/49 widths, yellow active target badge; ten source-matching help pastes consumed as close; Tab close/reopen and UI restart; preserved original drafts/PIDs; final explicit Enter. No whole-frame source/real-agent/IME/release claim.'}|ConvertTo-Json -Depth 12|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8;throw}
