$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/tab-order-20261123/desktop'
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
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'tab-order-20261123'
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
function TabRow {
 $text=StateText;$lines=$text -split "`r?`n"
 for($i=0;$i -lt $lines.Length-1;$i++){if($lines[$i].StartsWith('Agent Session Desk')){return $lines[$i+1]}}
 throw 'Product Tab row not observed'
}
function ExpectOrder($left,$right){$row=TabRow;$a=$row.IndexOf($left);$b=$row.IndexOf($right);if($a -lt 0 -or $b -lt 0 -or $a -ge $b){throw ('Unexpected Tab order: '+$row)}}
try{
 $null=WaitText 'READY Order A'
 Text 'Write-Output (''A:''+$PID) # []'
 $null=Shot 'a-draft-before-order-public'
 $null=InvokeVisual @{kind='dragTab';source='Order A';destination='Order B'}
 ExpectOrder 'Order B' 'Order A'
 $text=Shot 'a-draft-after-drag-public';if($text -match '(?m)^A:\d+\s*$'){throw 'Drag submitted A draft'}
 Key '^{PGUP}';$null=WaitText 'READY Order B'
 Text 'Write-Output (''B:''+$PID) # []'
 Key '{F6}';$null=WaitText '管理模式';Text ']';ExpectOrder 'Order A' 'Order B'
 $null=Shot 'b-managed-key-order-public'
 Key '{F6}';$null=WaitText '终端模式'
 $text=StateText;if($text -match '(?m)^B:\d+\s*$'){throw 'Management reorder submitted B draft'}
 Key '{ENTER}';$text=WaitText '(?m)^B:(\d+)\s*$';$bPid=[int]([regex]::Match($text,'(?m)^B:(\d+)\s*$').Groups[1].Value)
 Key '^{PGUP}';$null=WaitText 'READY Order A'
 Key '{F6}';$null=WaitText '管理模式';Text ']';ExpectOrder 'Order B' 'Order A'
 Key '{F6}';$null=WaitText '终端模式';$text=StateText
 if($text -notmatch 'Write-Output.*A:.*# \[\]' -or $text -match '(?m)^A:\d+\s*$'){throw 'Original A draft not preserved'}
 $null=Shot 'a-returned-original-draft-public'
 Key '{ENTER}';$text=WaitText '(?m)^A:(\d+)\s*$';$aPid=[int]([regex]::Match($text,'(?m)^A:(\d+)\s*$').Groups[1].Value)
 if($aPid -eq $bPid){throw 'Two shells have the same PID'}
 Key '{F6}';$null=WaitText '管理模式';Text 'x';$null=WaitText '(?m)^B:\d+\s*$'
 if((TabRow).Contains('Order A')){throw 'Active A Tab did not close'}
 $null=InvokeVisual @{kind='clickListRow';title='Order A'}
 Key '{ENTER}';$null=WaitText '终端模式';$null=WaitText '(?m)^A:\d+\s*$'
 Text 'Write-Output (''A2:''+$PID)';Key '{ENTER}';$text=WaitText '(?m)^A2:(\d+)\s*$'
 $again=[int]([regex]::Match($text,'(?m)^A2:(\d+)\s*$').Groups[1].Value);if($again -ne $aPid){throw 'Reopen replaced the original shell'}
 $null=Shot 'a-reopened-original-pid-public'
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit
 $base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$product=Join-Path $base 'order-product.exe';$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-tab-order','--state-dir',(Join-Path $testRoot 'state'))
 $before=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
 $after=((& $product @g list --json|Out-String)|ConvertFrom-Json)
 if($LASTEXITCODE -ne 0 -or @($after).Count -ne 2){throw 'Managed sessions not preserved'}
 foreach($v in $after){if(@($before|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id}).Count -ne 1){throw 'Original Session/Run changed'}}
 $shell=Join-Path $base 'pwsh7/pwsh.exe';$identities=@()
 foreach($number in @($aPid,$bPid)){$p=Get-Process -Id $number;if($p.Path -ne $shell){throw 'Shell image identity mismatch'};$identities+=,@{pid=$p.Id;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';commands=$script:sequence;before=$before;after=$after;shell_processes=$identities;a_pid=$aPid;b_pid=$bPid;a_reopened_pid=$again;checks=@('physical Tab drag','management bracket ordering','terminal literal brackets','drafts not submitted by reorder','navigation follows new order','close and reopen original Run','independent native shell PIDs')}|ConvertTo-Json -Depth 7|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{
 @{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8
 throw
}
