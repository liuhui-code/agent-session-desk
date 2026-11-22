$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/copy-final-20261123/desktop'
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
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$testRoot=Join-Path $base 'copy-final-20261123'
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
try{
 $null=WaitText 'COPY_READY Copy A'
 Text 'Write-Output (''A:''+$PID) # literal v []'
 $null=Shot 'a-original-draft-public'
 Key '{F6}';$null=WaitText '管理模式';Text 'v';$null=WaitText '滚动历史'
 Text '?';$null=WaitText '\(search up\)'
 Text (('x'*80)+'中文tail')
 $null=InvokeVisual @{kind='resize';width=600;height=539}
 $text=WaitText '\(search up\).*tail';$null=Shot 'search-long-prompt-63-public'
 Key '{ESC}';$null=WaitText '滚动历史'
 $null=InvokeVisual @{kind='resize';width=1113;height=657}
 Text '?';$null=WaitText '\(search up\)';Text 'SCROLL_NEEDLE_中文_α'
 $null=Shot 'search-unicode-prompt-public'
 Key '{ENTER}';$null=WaitText 'SCROLL_NEEDLE_中文_α'
 Text 'V';Start-Sleep -Milliseconds 1200;$null=Shot 'native-line-selected-public'
 Key '{ENTER}';$null=WaitText '已向终端发送复制请求'
 $null=InvokeVisual @{kind='expectClipboard';expected="SCROLL_NEEDLE_中文_α`n"}
 $text=WaitText 'Write-Output.*A:.*literal v \[\]'
 if($text -match 'A:(\d+) *│'){throw 'Search or copy submitted original A draft'}
 $null=Shot 'a-draft-after-copy-public'
 Key '^{PGDN}';$null=WaitText 'COPY_READY Copy B'
 Text 'Write-Output (''B:''+$PID) # untouched'
 Key '{F6}';$null=WaitText '管理模式';Text 'v';$null=WaitText '滚动历史'
 Text '?';$null=WaitText '\(search up\)';Text 'SHOULD_NOT_RUN'
 Key '{ESC}';$null=WaitText '滚动历史';Text 'q'
 $text=WaitText 'Write-Output.*B:.*untouched'
 if($text -match 'B:(\d+) *│' -or $text -match 'untouchedSHOULD_NOT_RUN'){throw 'Canceled search entered shell input'}
 Key '{ENTER}';$text=WaitText 'B:(\d+) *│';$bPid=[int]([regex]::Match($text,'B:(\d+) *│').Groups[1].Value)
 Key '^{PGUP}';$null=WaitText 'Write-Output.*A:.*literal v \[\]'
 Key '{ENTER}';$text=WaitText 'A:(\d+) *│';$aPid=[int]([regex]::Match($text,'A:(\d+) *│').Groups[1].Value)
 if($aPid -eq $bPid){throw 'Two shells have the same PID'}
 $null=Shot 'a-explicit-command-after-copy-public'
 Key '{F6}';$null=WaitText '管理模式';Text 'q';WaitExit
 $base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$product=Join-Path $base 'copy-final-product.exe';$g=@('--backend',(Join-Path $base 'backend/psmux.exe'),'--runtime-dir',(Join-Path $testRoot 'runtime'),'--namespace','asd-copy-final','--state-dir',(Join-Path $testRoot 'state'))
 $before=[IO.File]::ReadAllText((Join-Path $testRoot 'sessions.json'))|ConvertFrom-Json
 $after=((& $product @g list --json|Out-String)|ConvertFrom-Json)
 if($LASTEXITCODE -ne 0 -or @($after).Count -ne 2){throw 'Managed sessions not preserved'}
 foreach($v in $after){if(@($before|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id}).Count -ne 1){throw 'Original Session/Run changed'}}
 $shell=Join-Path $base 'pwsh7/pwsh.exe';$identities=@()
 foreach($number in @($aPid,$bPid)){$p=Get-Process -Id $number;if($p.Path -ne $shell){throw 'Shell identity mismatch'};$identities+=,@{pid=$p.Id;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}}
 $null=InvokeVisual @{kind='finish'}
 @{result='passed';commands=$script:sequence;before=$before;after=$after;shell_processes=$identities;a_pid=$aPid;b_pid=$bPid;checks=@('native copy-mode entry','long Unicode search prompt at 63 columns','search in backend history','native line selection','synthetic clipboard matches selected Unicode line','cancel leaves original shell draft','only explicit shell Enter executes','original Session/Run and independent shell PIDs')}|ConvertTo-Json -Depth 7|Set-Content (Join-Path $testRoot 'flow-result.json') -Encoding utf8
 Get-Content (Join-Path $testRoot 'flow-result.json') -Raw -Encoding utf8
}catch{
 @{result='failed';sequence=$script:sequence;error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $testRoot 'flow-failure.json') -Encoding utf8
 throw
}
