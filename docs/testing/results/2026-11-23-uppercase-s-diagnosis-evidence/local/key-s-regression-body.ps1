$script:sequence=162
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
try{
 if($capsOriginal.capsLock -ne $true){throw 'Original Caps Lock baseline differs'}
 $null=WaitText '管理模式'
 foreach($i in 1..3){Trial ('caps-on-physical-'+$i) $true}
 try{
  Mark 'caps-toggle-off';Key '{CAPSLOCK}';$capsMayHaveChanged=$true
  $off=Caps 'off';if($off.capsLock){throw 'Caps did not turn off'}
  foreach($i in 1..3){Trial ('caps-off-shift-'+$i) $false}
 }finally{
  if($capsMayHaveChanged){
   $current=Caps 'current'
   if($current.capsLock -ne $capsOriginal.capsLock){Mark 'caps-restore-original';Key '{CAPSLOCK}'}
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
