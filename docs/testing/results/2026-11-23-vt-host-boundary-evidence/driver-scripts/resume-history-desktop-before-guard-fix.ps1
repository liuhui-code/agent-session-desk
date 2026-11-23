$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'history-desktop-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'copy-final2-product.exe';$backend=Join-Path $base 'history-backend-candidate/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-history-desktop';$g=@('--backend',$backend,'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
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
 $o=@{label=$label;session_id=$s.session_id;run_id=$s.run_id;pane=$p.id;columns=$p.cols;rows=$p.rows;copy_mode=$p.copy_mode;alternate_screen=$p.alternate_screen;scroll_offset=$p.scroll_offset;view_offset=$p.view_offset;copy_cursor_row=$p.copy_cursor_row;copy_cursor_col=$p.copy_cursor_col;sel_start_row=$p.sel_start_row;sel_end_row=$p.sel_end_row;sel_mode=$p.sel_mode;screen_text=@($p.rows_v2|ForEach-Object{($_.runs.text -join '').TrimEnd()});status=$f.status_message}
 if([bool]$p.alternate_screen -ne [bool]$Alternate){throw 'Unexpected normal/alternate screen'}
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
$prior=[IO.File]::ReadAllText((Join-Path $testRoot 'ui-identities.json'))|ConvertFrom-Json;foreach($u in $prior){$script:uis.Add($u)}
$done=@(Get-ChildItem (Join-Path $root 'done') -Filter '*.json'|Sort-Object Name);if(!(Get-ScheduledTask AsdHistoryDesktop20261123).State.Equals('Running') -or @($sessions).Count -ne 2 -or @((Get-ChildItem (Join-Path $root 'commands') -Filter '*.json')).Count){throw 'Resume guards failed'};$script:sequence=[int]$done[-1].BaseName
try{
 $null=Observe 'initial-a' 'Copy A';$null=Observe 'initial-b' 'Copy B'
 Text 'Write-Output (''A:''+$PID) # literal v []';$null=Shot 'candidate-a-draft-public'
 Key '{F6}';$null=WaitText '管理模式';Text 'v';$null=WaitText '滚动历史';Text '?';$null=WaitText '\(search up\)';Text (('x'*80)+'中文tail')
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText '\(search up\).*tail';$null=Shot 'candidate-search-63-public';$null=Observe 'search-narrow' 'Copy A'
 Key '{ESC}';$null=WaitText '滚动历史';$null=InvokeVisual @{kind='resize';width=1113;height=657}
 Text '?';$null=WaitText '\(search up\)';Text 'SCROLL_NEEDLE_中文_α';Key '{ENTER}';$null=WaitText 'SCROLL_NEEDLE_中文_α'
 $null=Shot 'candidate-search-top-public';$null=Observe 'searched-wide' 'Copy A'
 Text 'V';$until=[DateTime]::UtcNow.AddSeconds(5);do{$selected=NativeFrame $sessions[0];if($null -ne $selected.layout.sel_start_row -and $selected.layout.sel_mode -eq 'line'){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until);if($selected.layout.sel_mode -ne 'line'){throw 'Native line selection not observed'};$null=Observe 'selected-line' 'Copy A';Start-Sleep -Milliseconds 300;$null=Shot 'candidate-line-selected-public';Key '{ENTER}';$null=WaitText '已向终端发送复制请求';$null=InvokeVisual @{kind='expectClipboard';expected="SCROLL_NEEDLE_中文_α`n"}
 $text=WaitText 'Write-Output.*A:.*literal v \[\]';if($text -match 'A:(\d+) *│'){throw 'Copy submitted A draft'};$null=Observe 'copied-live-a' 'Copy A'
 Key '^{PGDN}';$null=WaitText 'COPY_READY Copy B';Text 'Write-Output (''B:''+$PID) # untouched'
 Key '{F6}';$null=WaitText '管理模式';Text 'v';$null=WaitText '滚动历史';Text '?';$null=WaitText '\(search up\)';Text 'SHOULD_NOT_RUN';Key '{ESC}';$null=WaitText '滚动历史';Text 'q'
 $text=WaitText 'Write-Output.*B:.*untouched';if($text -match 'B:(\d+) *│' -or $text -match 'untouchedSHOULD_NOT_RUN'){throw 'Canceled search changed B draft'}
 Key '{ENTER}';$text=WaitText 'B:(\d+) *│';$bPid=[int]([regex]::Match($text,'B:(\d+) *│').Groups[1].Value)
 Key '^{PGUP}';$null=WaitText 'Write-Output.*A:.*literal v \[\]';Key '{ENTER}';$text=WaitText 'A:(\d+) *│';$aPid=[int]([regex]::Match($text,'A:(\d+) *│').Groups[1].Value)
 $ids=@(CheckShells);if($aPid -ne $ids[0].pid -or $bPid -ne $ids[1].pid){throw 'Observed shell PIDs changed'};$null=Shot 'candidate-explicit-a-public';$null=Observe 'explicit-a' 'Copy A';$null=Observe 'explicit-b' 'Copy B'
 # Continue through a controlled VT app before closing views or the manager.
 Paste ("& '"+$shell+"' -NoLogo -NoProfile -File .\alt-app.ps1");Key '{ENTER}';$null=WaitText 'VT_ALT_READY 中文 α';$null=Observe 'alt-wide' 'Copy A' -Alternate
 $null=InvokeVisual @{kind='resize';width=600;height=539};$null=WaitText 'VT_ALT_READY';$null=Shot 'candidate-alt-63-public';$null=Observe 'alt-narrow' 'Copy A' -Alternate
 Text 'ALT输入α';Key '{ENTER}';$null=WaitText 'VT_ALT_DONE';$null=Observe 'normal-after-alt' 'Copy A';$null=Shot 'candidate-after-alt-public'
 $alt=[IO.File]::ReadAllText((Join-Path $testRoot 'work/alt-observed.json'))|ConvertFrom-Json;if($alt.text -cne 'ALT输入α' -or $alt.pid -eq $aPid){throw 'Controlled VT app observation differs'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'VT_ALT_DONE';$null=Observe 'normal-restored-wide' 'Copy A'
 Key '{F6}';$null=WaitText '管理模式';Text 'x';$null=InvokeVisual @{kind='clickListRow';title='Copy A'};Key '{ENTER}';$null=WaitText 'VT_ALT_DONE';$null=CheckShells;$null=Observe 'tab-reopened-a' 'Copy A';$null=Shot 'candidate-tab-reopened-public'
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit
 $command="& '$product' --backend '$backend' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $command;Key '{ENTER}';$null=WaitText 'VT_ALT_DONE';RememberUI;$null=CheckShells;$null=Observe 'manager-reopened-a' 'Copy A';$null=Observe 'manager-reopened-b' 'Copy B';$null=Shot 'candidate-manager-reopened-public'
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit
 $after=((& $product @g list --json|Out-String)|ConvertFrom-Json);if($LASTEXITCODE -ne 0 -or $after.Count -ne 2){throw 'Managed sessions not preserved'}
 foreach($s in $after){if(@($sessions|Where-Object{$_.session_id -eq $s.session_id -and $_.run_id -eq $s.run_id}).Count -ne 1){throw 'Original Session/Run changed'}}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';commands=$script:sequence;before=$sessions;after=$after;shell_processes=@(CheckShells);ui_processes=$script:uis.ToArray();alt_observed=$alt;native_observations=$script:observations.Count;checks=@('copy and native history across 120-63-120','exact synthetic OSC52 clipboard','A/B original drafts and explicit submission','controlled alternate-screen app resize and Unicode line','view close/reopen','manager exit/reopen','original Session/Run and shell FILETIME')}|ConvertTo-Json -Depth 9|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8;throw}
