$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/visual-compare-20261123/desktop'
function Action($id,$value){
 $name=('{0:d3}.json' -f $id);$path=Join-Path $root "commands/$name";$temp=Join-Path $root "$name.tmp"
 [IO.File]::WriteAllText($temp,($value|ConvertTo-Json -Depth 5),[Text.UTF8Encoding]::new($false));Move-Item $temp $path
 $response=Join-Path $root "responses/$name";$until=[DateTime]::UtcNow.AddSeconds(15)
 while(!(Test-Path $response) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $response)){throw 'Driver response timeout'}
 $r=[IO.File]::ReadAllText($response)|ConvertFrom-Json
 if(!$r.ok){throw ($r|ConvertTo-Json -Depth 6)}
 return $r
}

$ready=Join-Path $root 'ready.json';$until=[DateTime]::UtcNow.AddSeconds(20)
while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if(!(Test-Path $ready)){throw 'Driver not ready'}
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'visual-compare-20261123'
$script:sequence=0
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
 do{$text=StateText;if($text -notmatch '(?m)^Agent Session Desk'){return};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
 throw 'Product UI has not exited; refusing to type a shell launch command'
}
function RefFrame($name,$width,$height){
 $null=InvokeVisual @{kind='resize';width=$width;height=$height}
 [IO.File]::WriteAllText((Join-Path $testRoot 'frame-request.txt'),$name,[Text.UTF8Encoding]::new($false))
 $until=[DateTime]::UtcNow.AddSeconds(12);$result=$null
 do{
  $file=Join-Path $testRoot 'frame-receipt.json'
  if(Test-Path $file){try{$result=[IO.File]::ReadAllText($file)|ConvertFrom-Json}catch{$result=$null}}
  if($result -and $result.frame -eq $name -and $result.ok){break}
  Start-Sleep -Milliseconds 100
 }while([DateTime]::UtcNow -lt $until)
 if(!$result -or !$result.ok -or $result.frame -ne $name){throw ('Reference replay failed: '+($result|ConvertTo-Json -Compress))}
 Start-Sleep -Milliseconds 300
 $null=Shot ('reference-replay-'+$name+'-public')
}
function LaunchProduct($phase,$executable){
 WaitExit
 $product=Join-Path $base $executable;$backend=Join-Path $base 'backend/psmux.exe'
 $command="& '$product' --backend '$backend' --runtime-dir '$(Join-Path $testRoot 'runtime')' --namespace asd-visual-compare --state-dir '$(Join-Path $testRoot 'state')'"
 Paste $command;Key '{ENTER}';$null=WaitText 'Plain terminal output'
 $null=Shot ($phase+'-focused-120-public')
 Key '{F6}';$null=WaitText '管理模式';$null=Shot ($phase+'-manager-120-public')
 Text 'S';$null=WaitText '侧栏密度';$null=Shot ($phase+'-settings-120-public')
 Key '{ESC}';$null=WaitText '管理模式'
 Text 'n';$null=WaitText '新建终端';$null=Shot ($phase+'-new-dialog-120-public')
 Key '{ESC}';$null=WaitText '管理模式'
 $null=InvokeVisual @{kind='resize';width=600;height=520};$null=Shot ($phase+'-manager-63-public')
 $null=InvokeVisual @{kind='resize';width=400;height=520};$null=Shot ($phase+'-manager-49-public')
 Text 'S';$null=WaitText '侧栏密度';$null=Shot ($phase+'-settings-49-public')
 Key '{ESC}';$null=WaitText '管理模式'
 $null=InvokeVisual @{kind='resize';width=1113;height=657}
 if($phase -eq 'after'){
  Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText 'Plain terminal output'
  Text 'alpha';Key '^{PGDN}';$null=WaitText '│Review B\s*│'
  Text 'beta';Key '^{PGUP}';$null=WaitText '> alpha'
  $null=Shot 'after-returned-a-draft-public'
  if((StateText) -match 'echo: alpha'){throw 'Switching implicitly submitted A draft'}
  Key '{ENTER}';$null=WaitText 'echo: alpha'
  Key '^{PGDN}';$null=WaitText '> beta';Key '{ENTER}';$null=WaitText 'echo: beta'
  $null=Shot 'after-echo-b-public'
  Key '{F6}';$null=WaitText '管理模式'
 }
 Text 'q';WaitExit
}
try{
 $ready=Join-Path $testRoot 'replay-ready.json';$until=[DateTime]::UtcNow.AddSeconds(20)
 while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $ready)){throw 'Reference replayer did not become ready'}
 RefFrame 'upstream-focused-120' 1113 657
 RefFrame 'upstream-focused-wide-120' 1113 657
 RefFrame 'upstream-manager-wide-120' 1113 657
 RefFrame 'upstream-settings-120' 1113 657
 RefFrame 'upstream-new-dialog-120' 1113 657
 RefFrame 'upstream-manager-63' 600 520
 RefFrame 'upstream-manager-49' 400 520
 RefFrame 'upstream-settings-49' 400 520
 [IO.File]::WriteAllText((Join-Path $testRoot 'replay.stop'),'stop',[Text.UTF8Encoding]::new($false))
 $finished=Join-Path $testRoot 'replay-finished.json';$until=[DateTime]::UtcNow.AddSeconds(10)
 while(!(Test-Path $finished) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $finished)){throw 'Replayer did not exit normally'}
 $null=InvokeVisual @{kind='resize';width=1113;height=657}
 LaunchProduct 'before' 'product-before.exe'
 LaunchProduct 'after' 'product-after.exe'
 $product=Join-Path $base 'product-after.exe';$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-visual-compare','--state-dir',(Join-Path $testRoot 'state'))
 $before=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
 $after=((& $product @g list --json|Out-String)|ConvertFrom-Json)
 if($LASTEXITCODE -ne 0 -or @($after).Count -ne 2){throw 'Managed sessions not preserved'}
 foreach($v in $after){if(@($before|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id}).Count -ne 1){throw 'Original Session/Run changed'}}
 $fixture=Join-Path $base 'visual-fixture.exe'
 $processes=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $fixture})
 if($processes.Count -ne 2){throw 'Cannot independently account for both fixture processes'}
 $identities=@();foreach($item in $processes){$p=Get-Process -Id $item.ProcessId;if($p.Path -ne $fixture){throw 'Fixture identity changed'};$identities+=,@{pid=$p.Id;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';commands=$script:sequence;before=$before;after=$after;fixture_processes=$identities;checks=@('original upstream frames replayed with exact console dimension checks','same Windows Terminal profile and Consolas 12 font','actual before and after product views','settings/new-dialog widths','wide stacked single-list views','A-B-A drafts and explicit echo','original Session/Run and two live fixtures preserved')}|ConvertTo-Json -Depth 7|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{
 @{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8
 throw
}
