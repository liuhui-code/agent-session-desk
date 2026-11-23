$script:sequence=62
$baseline=[IO.File]::ReadAllText((Join-Path $testRoot 'baseline-result.json'))|ConvertFrom-Json
foreach($id in $baseline.ui_processes){$script:uis.Add($id)}
foreach($o in ([IO.File]::ReadAllText((Join-Path $testRoot 'native-observations.json'))|ConvertFrom-Json)){$script:observations.Add($o);$script:panes[$o.session_id]=$o.pane}
$trials=[Collections.Generic.List[object]]::new()
function Mark([string]$phase){[IO.File]::WriteAllText((Join-Path $testRoot 'diag-step.txt'),$phase,[Text.UTF8Encoding]::new($false))}
try{
 if((Get-ScheduledTask 'AsdSidebarKeyDesktop20261123').State -ne 'Running'){throw 'Original desktop task not live'}
 $null=WaitText '管理模式';Text 'q';WaitExit
 $product=Join-Path $base 'asd-key-s-instrumented.exe'
 if((Get-FileHash $product).Hash.ToLowerInvariant() -cne '75a052d5be803cba84a798be8b4a6367470d57de34ac383be1b0bcbb599470bc'){throw 'Instrumented binary differs'}
 Mark 'debug-launch'
 $launch="`$env:ASD_KEY_DIAG_ROOT='$testRoot'; & '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace $ns --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $launch;Key '{ENTER}';$null=WaitText 'COPY_READY Copy A';RememberUI;Key '{F6}';$null=WaitText '管理模式'
 foreach($kind in @('sendkeys','unicode','held-shift')){
  foreach($i in 1..3){
   $phase=$kind+'-'+$i;Mark $phase
   switch($kind){'sendkeys'{Key 'S'};'unicode'{Text 'S'};'held-shift'{$null=InvokeVisual @{kind='shiftS'}}}
   $opened=$false;$clock=[Diagnostics.Stopwatch]::StartNew()
   do{$text=StateText;if($text -match '侧栏密度'){$opened=$true;break};Start-Sleep -Milliseconds 100}while($clock.Elapsed.TotalSeconds -lt 2)
   if($i -eq 1){$null=Shot ($phase+'-public')}
   $trials.Add(@{phase=$phase;opened_settings=$opened;last_sequence=$script:sequence;observed_text=$text})
   Mark ('cleanup-'+$phase);Key '{ESC}';$null=WaitText '管理模式';$null=WaitText 'COPY_READY Copy A'
  }
 }
 Mark 'debug-observe';$o=Observe 'debug-no-terminal-input' 'Copy A'
 if(($o.capture_text -split '\r?\n'|Where-Object{$_.Trim()}|Select-Object -Last 1).Trim() -cne '>'){throw 'Debug shortcut test changed target input'}
 @{result='observed';utc=[DateTime]::UtcNow.ToString('o');sequence=$script:sequence;trials=$trials.ToArray();ui_processes=$script:uis.ToArray();shells=@(CheckShells);instrumented_sha256=(Get-FileHash $product).Hash;scope='Same original product with log-only build overlays; raw KEY_EVENT and decoded Bubble Tea key boundaries; no dependency upgrade'}|ConvertTo-Json -Depth 8|Set-Content (Join-Path $testRoot 'probe-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'probe-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'probe-failure.json') -Encoding utf8;throw}
