$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'sidebar-key-diagnosis-20261123';$root=Join-Path $testRoot 'desktop'
$product=Join-Path $base 'asd-sidebar-key-diag.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$ns='asd-sidebar-key-diag';$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace',$ns,'--state-dir',(Join-Path $testRoot 'state'))
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
 $ids=@();foreach($name in @('Copy_A')){$i=[IO.File]::ReadAllText((Join-Path $testRoot ('work/'+$name+'.identity.json')))|ConvertFrom-Json;$p=Get-Process -Id $i.pid;if($p.Path -ne $i.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $i.filetime){throw 'Original shell changed'};$ids+=,$i};return $ids
}
function CheckDraft([string]$label,[string]$title,[string]$draft,[string]$output){
 $o=Observe $label $title
 if($o.capture_text -notmatch [regex]::Escape($draft) -or $o.capture_text -match ('(?m)^'+$output+':\d+\s*$')){throw 'Original draft was changed or submitted before explicit Enter'}
}
$script:sequence=180
$probe=[IO.File]::ReadAllText((Join-Path $testRoot 'probe-result.json'))|ConvertFrom-Json
foreach($id in $probe.ui_processes){$script:uis.Add($id)}
foreach($o in ([IO.File]::ReadAllText((Join-Path $testRoot 'native-observations.json'))|ConvertFrom-Json)){$script:observations.Add($o);$script:panes[$o.session_id]=$o.pane}
$trials=[Collections.Generic.List[object]]::new();$capsOriginal=[IO.File]::ReadAllText((Join-Path $testRoot 'caps-before.json'))|ConvertFrom-Json;$capsMayHaveChanged=$false;$capsRestored=$null
function Mark([string]$phase){[IO.File]::WriteAllText((Join-Path $testRoot 'diag-step.txt'),$phase,[Text.UTF8Encoding]::new($false))}
function Caps([string]$phase){
 $null=& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base ('query-caps-'+$phase+'.ps1')))))
 $c=[IO.File]::ReadAllText((Join-Path $testRoot ('caps-'+$phase+'.json')))|ConvertFrom-Json
 if($c.result -ne 'observed' -or $c.capsLock -isnot [bool]){throw 'Caps state unavailable'};return $c
}
function UppercaseS([bool]$caps){
 if($caps){$null=InvokeVisual @{kind='physical';text='S'}}else{$null=InvokeVisual @{kind='shiftS'}}
}
function Trial([string]$phase,[bool]$caps){
 Mark $phase;UppercaseS $caps;$null=WaitText '侧栏密度'
 if($phase.EndsWith('-1')){$null=Shot ($phase+'-public')}
 $trials.Add(@{phase=$phase;capsLock=$caps;used_shift=(!$caps);opened_settings=$true;sequence=$script:sequence})
 Mark ('cleanup-'+$phase);Key '{ESC}';$null=WaitText '管理模式';$null=WaitText 'COPY_READY Copy A'
}
foreach($row in @(@(1,164,165),@(2,170,171),@(3,175,176))){
 $sent=[IO.File]::ReadAllText((Join-Path $root ('responses/{0:d3}.json' -f $row[1])))|ConvertFrom-Json
 $observed=[IO.File]::ReadAllText((Join-Path $root ('responses/{0:d3}.json' -f $row[2])))|ConvertFrom-Json
 if(!$sent.ok -or $sent.command.kind -ne 'physical' -or $sent.command.text -cne 'S' -or !$observed.ok -or $observed.state.text -notmatch '侧栏密度'){throw 'Prior green trial cannot be independently reconstructed'}
 $trials.Add(@{phase=('caps-on-physical-'+$row[0]);capsLock=$true;used_shift=$false;opened_settings=$true;sequence=$row[2];reconstructed_from_actual_receipts=$true})
}
[IO.File]::Copy((Join-Path $testRoot 'regression-failure.json'),(Join-Path $testRoot 'regression-failure-sendkeys-capslock.json'))
try{
 $null=WaitText '管理模式'
 try{
  Mark 'caps-toggle-off';$null=InvokeVisual @{kind='physical';text=([string][char]20)};$capsMayHaveChanged=$true
  $off=Caps 'off';if($off.capsLock){throw 'Caps did not turn off'}
  foreach($i in 1..3){Trial ('caps-off-shift-'+$i) $false}
 }finally{
  if($capsMayHaveChanged){
   $current=Caps 'current'
   if($current.capsLock -ne $capsOriginal.capsLock){Mark 'caps-restore-original';$null=InvokeVisual @{kind='physical';text=([string][char]20)}}
   $capsRestored=Caps 'final';if($capsRestored.capsLock -ne $capsOriginal.capsLock){throw 'Original Caps Lock not restored'};$capsMayHaveChanged=$false
  }
 }
 $product=Join-Path $base 'asd-key-s-instrumented.exe';Mark 'debug-quit';Text 'q';WaitExit
 $product=Join-Path $base 'asd-sidebar-key-diag.exe'
 if((Get-FileHash $product).Hash.ToLowerInvariant() -cne 'f58e80afbef817bfbf7e42528a57607a4e664bf068e2dd417ff1f6b4a7acf277'){throw 'Original product differs'}
 $launch="Remove-Item Env:ASD_KEY_DIAG_ROOT -ErrorAction SilentlyContinue; & '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $launch;Key '{ENTER}';$null=WaitText 'COPY_READY Copy A';RememberUI;Key '{F6}';$null=WaitText '管理模式'
 foreach($i in 1..3){Trial ('original-caps-aware-'+$i) $true}
 Key '{F6}';$null=WaitText '终端模式';UppercaseS $true;$null=WaitText '> S';$null=Shot 'original-terminal-uppercase-S-public'
 $o=Observe 'original-terminal-uppercase-S' 'Copy A';if(($o.capture_text -split '\r?\n'|Where-Object{$_.Trim()}|Select-Object -Last 1).Trim() -cne '> S'){throw 'Original terminal did not receive exactly uppercase S'}
 Key '{F6}';$null=WaitText '管理模式';UppercaseS $true;$null=WaitText '侧栏密度';$null=Shot 'original-settings-with-draft-public'
 $o=Observe 'original-settings-draft-preserved' 'Copy A';if(($o.capture_text -split '\r?\n'|Where-Object{$_.Trim()}|Select-Object -Last 1).Trim() -cne '> S'){throw 'Settings changed original S draft'}
 Key '{ESC}';$null=WaitText '管理模式';$null=WaitText 'COPY_READY Copy A';Text 'q';WaitExit
 $o=Observe 'original-after-UI-quit-draft' 'Copy A';if(($o.capture_text -split '\r?\n'|Where-Object{$_.Trim()}|Select-Object -Last 1).Trim() -cne '> S'){throw 'UI quit lost or submitted S draft'}
 $afterCaps=Caps 'after-control';if($afterCaps.capsLock -ne $capsOriginal.capsLock){throw 'Caps changed during original control'}
 $after=(& $product @g list --json|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or $after.session_id -ne $sessions.session_id -or $after.run_id -ne $sessions.run_id){throw 'Original Run changed'}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';utc=[DateTime]::UtcNow.ToString('o');sequence=$script:sequence;trials=$trials.ToArray();caps_before=$capsOriginal;caps_off=$off;caps_restored=$capsRestored;caps_after_control=$afterCaps;ui_processes=$script:uis.ToArray();before=$sessions;after=$after;before_processes=$probe.shells;after_processes=@(CheckShells);scope='Caps-aware physical uppercase S; raw and decoded boundaries, original binary regression and terminal input isolation; no product fix or layout installation'}|ConvertTo-Json -Depth 10|Set-Content (Join-Path $testRoot 'key-s-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'key-s-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;caps_may_have_changed=$capsMayHaveChanged;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'regression-failure.json') -Encoding utf8;throw}
