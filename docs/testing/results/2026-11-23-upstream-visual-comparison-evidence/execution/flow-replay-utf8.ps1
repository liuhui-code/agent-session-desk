$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/visual-replay-utf8-20261123/desktop'
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
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'visual-replay-utf8-20261123'
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
try{
 $ready=Join-Path $testRoot 'replay-ready.json';$until=[DateTime]::UtcNow.AddSeconds(20)
 while(!(Test-Path $ready) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $ready)){throw 'Reference replayer did not become ready'}
 $v=[IO.File]::ReadAllText($ready)|ConvertFrom-Json;if($v.output_code_page -ne 65001){throw 'UTF-8 output code page not active'}
 RefFrame 'upstream-focused-120' 1113 657
 $screen=StateText;if(!$screen.Contains('中文') -or !$screen.Contains('╭')){throw 'Original Unicode content not rendered'}
 RefFrame 'upstream-focused-wide-120' 1113 657
 RefFrame 'upstream-manager-wide-120' 1113 657
 RefFrame 'upstream-settings-120' 1113 657
 RefFrame 'upstream-new-dialog-120' 1113 657
 RefFrame 'upstream-manager-63' 600 539
 RefFrame 'upstream-manager-49' 400 539
 RefFrame 'upstream-settings-49' 400 539
 [IO.File]::WriteAllText((Join-Path $testRoot 'replay.stop'),'stop',[Text.UTF8Encoding]::new($false))
 $finished=Join-Path $testRoot 'replay-finished.json';$until=[DateTime]::UtcNow.AddSeconds(10)
 while(!(Test-Path $finished) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
 if(!(Test-Path $finished)){throw 'Replayer did not exit normally'}
 $null=InvokeVisual @{kind='finish'}
 @{result='captured';commands=$script:sequence;utf8_code_page=65001;row_validation='pending independent local comparison'}|ConvertTo-Json|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{
 @{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8
 throw
}finally{[IO.File]::WriteAllText((Join-Path $testRoot 'replay.stop'),'stop',[Text.UTF8Encoding]::new($false))}
