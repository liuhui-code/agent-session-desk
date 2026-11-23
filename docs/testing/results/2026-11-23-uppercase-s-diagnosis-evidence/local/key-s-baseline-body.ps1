$script:trials=[Collections.Generic.List[object]]::new()
try{
 $ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20);while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100};if(!(Test-Path $ready)){throw 'Key desktop not ready'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657};$null=WaitText 'COPY_READY Copy A';RememberUI;$before=@(CheckShells)
 Key '{F6}';$null=WaitText '管理模式'
 foreach($i in 1..3){
  Key 'S';$opened=$false;$clock=[Diagnostics.Stopwatch]::StartNew()
  do{$text=StateText;if($text -match '侧栏密度'){$opened=$true;break};Start-Sleep -Milliseconds 100}while($clock.Elapsed.TotalSeconds -lt 2)
  $null=Shot ('baseline-sendkeys-'+$i+'-public');$sendKeysSequence=$script:sequence
  Key '{ESC}';$null=WaitText '管理模式';$null=WaitText 'COPY_READY Copy A'
  Text 'S';$null=WaitText '侧栏密度';$null=Shot ('baseline-unicode-'+$i+'-public')
  Key '{ESC}';$null=WaitText '管理模式';$null=WaitText 'COPY_READY Copy A'
  $script:trials.Add(@{trial=$i;sendkeys_opened_settings=$opened;sendkeys_wait_milliseconds=$clock.ElapsedMilliseconds;unicode_opened_settings=$true;sendkeys_capture_sequence=$sendKeysSequence})
 }
 $o=Observe 'baseline-no-terminal-input' 'Copy A'
 if(($o.capture_text -split '\r?\n'|Where-Object{$_.Trim()}|Select-Object -Last 1).Trim() -cne '>'){throw 'Baseline shortcut tests changed terminal input'}
 @{result='observed';utc=[DateTime]::UtcNow.ToString('o');sequence=$script:sequence;trials=$script:trials.ToArray();ui_processes=$script:uis.ToArray();shells=$before;scope='Unmodified original product, same keyboard layout, SendKeys versus Unicode S; no language/caps changes'}|ConvertTo-Json -Depth 8|Set-Content (Join-Path $testRoot 'baseline-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'baseline-result.json') -Raw -Encoding utf8
}catch{@{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'baseline-failure.json') -Encoding utf8;throw}
