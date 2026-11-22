$ErrorActionPreference='Stop';$PSNativeCommandUseErrorActionPreference=$false
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'm1-history-ui'
if(Test-Path $root){throw 'Fresh test directory required'}
New-Item -ItemType Directory -Path $root|Out-Null
$product=Join-Path $base 'asd-history.exe';$psmux=Join-Path $base 'psmux.exe'
if(!(Test-Path $psmux)){$psmux=Join-Path $base 'psmux/psmux.exe'}
if(!(Test-Path $psmux)){$psmux=(Get-ChildItem $base -Filter psmux.exe -Recurse|Select-Object -First 1).FullName}
if((Get-FileHash $psmux).Hash.ToLowerInvariant() -ne 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'){throw 'psmux pin mismatch'}
$argsGlobal=@('--backend',$psmux,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-history-ui')
$commands=[Collections.Generic.List[object]]::new();$owned=[Collections.Generic.List[string]]::new()
function Product([string[]]$a){$errPath=Join-Path $root "stderr-$($commands.Count).txt";$text=(& $product @argsGlobal @a 2>$errPath|Out-String);$code=$LASTEXITCODE;$commands.Add(@{args=$a;code=$code;stdout=$text;stderr=[IO.File]::ReadAllText($errPath)});if($code -ne 0){throw "CLI failed $text"};return $text}
$failure=$null
try{
 foreach($test in @('ui','manager')){
  $testPath=Join-Path $base "$test-history-tests.exe"
  $text=(& $testPath -test.v 2>&1|Out-String)
  [IO.File]::WriteAllText((Join-Path $root "$test-tests.log"),$text)
  if($LASTEXITCODE -ne 0 -or $text.Contains('SKIP')){throw "$test tests failed or skipped"}
 }
 foreach($n in @(1,2)){
  $a=@('start','--name',"history-fixture-$n",'--env',"ASD_FIXTURE_REPORT_PATH=$(Join-Path $root "fixture-$n.jsonl")")
  if($n -eq 1){$a+=@('--env','ASD_FIXTURE_EXIT_AFTER_START=1','--env','ASD_FIXTURE_EXIT_CODE=7')}
  $s=Product ($a+@('--',(Join-Path $base 'asd-fixture.exe'),"synthetic-history-$n"))|ConvertFrom-Json
  $owned.Add($s.session_id)
  if($n -eq 1){
   $until=[DateTime]::UtcNow.AddSeconds(20);$dead=$false
   do{$rows=@((Product @('list','--json'))|ConvertFrom-Json);$row=@($rows|Where-Object{$_.session_id -eq $s.session_id});$dead=$row.Count -eq 1 -and $row[0].terminal_state -eq 'exited';if(!$dead){Start-Sleep -Milliseconds 100}}while(!$dead -and [DateTime]::UtcNow -lt $until)
   if(!$dead){throw 'exit observation absent'}
  }
  $null=Product @('stop',$s.session_id);$owned.Remove($s.session_id)|Out-Null
 }
 $history=Product @('history');[IO.File]::WriteAllText((Join-Path $root 'history.json'),$history)
 $h=@($history|ConvertFrom-Json)
 if($h.Count -ne 2 -or @($h|Where-Object{$_.stop_confirmed_at}).Count -ne 2){throw 'history setup incomplete'}
 $list=Product @('list','--json');if(@($list|ConvertFrom-Json).Count -ne 0){throw 'owned sessions not cleaned'}
 $command="& '$product' --backend '$psmux' --runtime-dir '$(Join-Path $root 'runtime')' --namespace asd-history-ui"
 [IO.File]::WriteAllText((Join-Path $root 'launch.txt'),$command)
} catch {$failure=$_.Exception.Message} finally{
 foreach($id in $owned.ToArray()){try{$null=Product @('stop',$id)}catch{}}
 $commands|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'commands.json') -Encoding utf8
 @{utc=[DateTime]::UtcNow.ToString('o');os=[Environment]::OSVersion.VersionString;product_sha256=(Get-FileHash $product).Hash.ToLowerInvariant();psmux=$psmux;psmux_sha256=(Get-FileHash $psmux).Hash.ToLowerInvariant();ui_tests_sha256=(Get-FileHash (Join-Path $base 'ui-history-tests.exe')).Hash.ToLowerInvariant();manager_tests_sha256=(Get-FileHash (Join-Path $base 'manager-history-tests.exe')).Hash.ToLowerInvariant();failure=$failure}|ConvertTo-Json|Set-Content (Join-Path $root 'setup-result.json') -Encoding utf8
}
if($failure){throw $failure}
'Native tests and synthetic history setup passed.'
