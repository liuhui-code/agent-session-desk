$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'help-parity-desktop-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'asd-help-parity.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-help-parity';$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
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
function Draft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title;$lines=@(($o.capture_text -split '\r?\n')|Where-Object{$_.Trim()})
 if(!$lines.Count -or $lines[-1].Trim() -cne ('> '+$draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw ('Pending original draft differs: '+$label)}
}
$originalSessions=@($sessions)
function CurrentSessions{$rows=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or @($rows).Count -ne 2){throw 'Live set changed'};foreach($s in $originalSessions){if(@($rows|Where-Object{$_.session_id -ceq $s.session_id -and $_.run_id -ceq $s.run_id -and $_.backend_id -ceq $s.backend_id -and $_.command -ceq $s.command}).Count -ne 1){throw 'Original identity changed'}};return $rows}
function Tail{$text=WaitText '关闭 Tab 不停止 Run';if($text -match '下方还有'){throw 'Help tail still claims more rows'}}
$a="Write-Output ('HELP_A:'+`$PID) # ?nq";$b="Write-Output ('HELP_B:'+`$PID)"
$prefPath=Join-Path $testRoot 'state/ui.toml';$prefHash=(Get-FileHash $prefPath).Hash
try{
 $null=WaitText 'COPY_READY Copy A';RememberUI;$before=@(CheckShells)
 Paste "Write-Output ('HELP_A:'+`$PID) # ";Text '?nq';Draft 'A-initial-terminal-letters' 'Copy A' $a 'HELP_A'
 Key '^{PGDN}';$null=WaitText 'COPY_READY Copy B';Paste $b;Draft 'B-initial' 'Copy B' $b 'HELP_B';Key '^{PGUP}';$null=WaitText ([regex]::Escape($a));Key '{F6}';$null=WaitText '管理模式'
 Text '?';$null=WaitText '五键入门';$null=Shot 'help-short-120-public'
 Key '{TAB}';$null=WaitText '全部快捷键';$null=Shot 'help-full-first-120-public'
 Key '{END}';Tail;$null=Shot 'help-full-last-120-public';$tail=StateText;Key '{PGDN}';if((StateText) -cne $tail){throw 'Help scrolled beyond end'}
 Key '{HOME}';$first=WaitText '全部快捷键';Text 'j';$down=StateText;if($down -ceq $first -or $down -notmatch '下方还有'){throw 'j did not scroll help'};Text 'k';if((StateText) -cne $first){throw 'k did not return to first row'}
 Key '{PGDN}';$page=StateText;if($page -ceq $first){throw 'PageDown did not scroll'};Key '{PGUP}';if((StateText) -cne $first){throw 'PageUp did not return'}
 Key '^d';if((StateText) -ceq $first){throw 'Ctrl+D did not scroll'};Key '^u';if((StateText) -cne $first){throw 'Ctrl+U did not return'}
 Text 'G';Tail;Text 'g';$null=WaitText '全部快捷键';Key '{DOWN}';Key '{UP}';if((StateText) -cne $first){throw 'Arrow navigation changed the first page'}
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '全部快捷键';$null=Shot 'help-full-first-63-public';Key '{END}';Tail;$null=Shot 'help-full-last-63-public'
 $null=InvokeVisual @{kind='resize';width=400;height=539};Key '{END}';Tail;$null=Shot 'help-full-last-49-public';Key '{HOME}';$null=WaitText '全部快捷键';$null=Shot 'help-full-first-49-public'
 Key '{TAB}';$null=WaitText '五键入门';$null=Shot 'help-short-49-public';Paste 'MUST_NOT_REACH_THE_SHELL';$null=WaitText '五键入门';Text 'n';$null=WaitText '管理模式';$null=@(CurrentSessions)
 Draft 'A-after-n-close' 'Copy A' $a 'HELP_A';Draft 'B-after-n-close' 'Copy B' $b 'HELP_B'
 Text '?';$null=WaitText '五键入门';Key '^{PGDN}';$null=WaitText '管理模式';Key '{F6}';$null=WaitText ([regex]::Escape($a));Key '{F6}';$null=WaitText '管理模式'
 Text '?';$null=WaitText '五键入门';Key '{ENTER}';$null=WaitText '管理模式';Draft 'A-after-enter-close' 'Copy A' $a 'HELP_A'
 Text '?';$null=WaitText '五键入门';Key '^c';$null=WaitText '管理模式';$null=@(CurrentSessions)
 Text '?';$null=WaitText '五键入门';Key '{TAB}';$null=WaitText '全部快捷键';$null=InvokeVisual @{kind='resize';width=400;height=300};$null=Shot 'help-full-short-height-public'
 $null=InvokeVisual @{kind='resize';width=1113;height=657};Key '{TAB}';$null=WaitText '五键入门';Text 'q';$null=WaitText '管理模式'
 if((Get-FileHash $prefPath).Hash -cne $prefHash){throw 'Help changed UI preferences'}
 Draft 'A-after-help' 'Copy A' $a 'HELP_A';Draft 'B-after-help' 'Copy B' $b 'HELP_B'
 Text 'q';WaitExit;$null=@(CurrentSessions);Draft 'A-after-quit' 'Copy A' $a 'HELP_A';Draft 'B-after-quit' 'Copy B' $b 'HELP_B'
 $launch="& '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $launch;Key '{ENTER}';$null=WaitText '终端模式';RememberUI;Key '{F6}';$null=WaitText '管理模式';Text '?';$null=WaitText '五键入门';$null=Shot 'help-reopened-short-120-public';Key '{ESC}';$null=WaitText '管理模式'
 if((Get-FileHash $prefPath).Hash -cne $prefHash){throw 'Help reopen persisted temporary page state'}
 Text '/';$null=WaitText '搜索';Text 'Copy A';Key '{ENTER}';$null=WaitText '终端模式';Draft 'A-before-explicit-enter' 'Copy A' $a 'HELP_A';Key '{ENTER}';$null=WaitText ('HELP_A:'+$before[0].pid);$oa=Observe 'A-explicit-enter' 'Copy A';if(@([regex]::Matches($oa.capture_text,('(?m)^HELP_A:'+$before[0].pid+'\s*$'))).Count -ne 1){throw 'A executed elsewhere/more than once'}
 Key '^{PGDN}';$null=WaitText '终端模式';Draft 'B-before-explicit-enter' 'Copy B' $b 'HELP_B';Key '{ENTER}';$null=WaitText ('HELP_B:'+$before[1].pid);$ob=Observe 'B-explicit-enter' 'Copy B';if(@([regex]::Matches($ob.capture_text,('(?m)^HELP_B:'+$before[1].pid+'\s*$'))).Count -ne 1){throw 'B executed elsewhere/more than once'}
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit;$after=@(CurrentSessions);$null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');commands=$script:sequence;before=$originalSessions;after=$after;before_processes=$before;after_processes=@(CheckShells);ui_processes=$script:uis.ToArray();preferences_before=$prefHash;preferences_after=(Get-FileHash $prefPath).Hash;native_captures=$script:observations.Count;scope='Two ordinary PowerShells; short/full help, navigation aliases and final page, 120/63/49/short height, literal terminal ?nq and modal paste/key isolation, unchanged preferences, original drafts/PIDs, UI restart and final explicit Enter. No real-agent/IME/general release claim.'}|ConvertTo-Json -Depth 12|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8;throw}
