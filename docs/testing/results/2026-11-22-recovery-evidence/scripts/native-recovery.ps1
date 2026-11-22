$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';$PSNativeCommandUseErrorActionPreference=$false
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'm1-recovery-run1'
if(Test-Path $root){throw 'Fresh test directory required'}
New-Item -ItemType Directory $root|Out-Null
$product=Join-Path $base 'asd-recovery.exe';$psmux=Join-Path $base 'backend/psmux.exe';$fixture=Join-Path $base 'asd-fixture.exe'
if((Get-FileHash $psmux).Hash.ToLowerInvariant() -ne 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'){throw 'psmux pin mismatch'}
$checks=[Collections.Generic.List[object]]::new();$commands=[Collections.Generic.List[object]]::new();$owned=$null;$failure=$null
$globalArgs=@('--backend',$psmux,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-recovery-native','--state-dir',(Join-Path $root 'state'))
function Product([string[]]$g,[string[]]$a,[bool]$success=$true){
 $stderr=Join-Path $root "stderr-$($commands.Count).txt"
 $text=(& $product @g @a 2>$stderr|Out-String);$code=$LASTEXITCODE
 $commands.Add(@{global=$g;args=$a;code=$code;stdout=$text;stderr=[IO.File]::ReadAllText($stderr)})
 if(($success -and $code -ne 0) -or (!$success -and $code -eq 0)){throw "Unexpected CLI outcome $code : $text"}
 return $text
}
try{
 foreach($name in @('store','cli')) {
  $exe=Join-Path $base "$name-recovery-tests.exe"
  $text=(& $exe '-test.v' '-test.timeout=90s' 2>&1|Out-String);$code=$LASTEXITCODE
  [IO.File]::WriteAllText((Join-Path $root "$name-tests.log"),$text)
  if($code -ne 0 -or $text.Contains('SKIP')){throw "$name native tests failed or skipped"}
  $checks.Add(@{scenario="$name native tests";sha256=(Get-FileHash $exe).Hash;result='passed'})
 }
 $cwd=Join-Path $root '工作 目录';New-Item -ItemType Directory $cwd|Out-Null
 $config=Join-Path $root 'agents.toml';$report=Join-Path $root 'fixture.jsonl'
 $toml="schema_version = 1`n[templates.check]`nname = 'recovery fixture'`nexecutable = '$fixture'`nargs = ['recovery-marker']`ncwd = '$cwd'`n[templates.check.env]`nASD_FIXTURE_REPORT_PATH = '$report'`nASD_FIXTURE_EXIT_AFTER_START = '1'`nASD_FIXTURE_EXIT_CODE = '7'`n"
 [IO.File]::WriteAllText($config,$toml,[Text.UTF8Encoding]::new($false))
 $start=Product $globalArgs @('--config',$config,'start','--template','check')|ConvertFrom-Json;$owned=$start.session_id
 $deadline=[DateTime]::UtcNow.AddSeconds(15);$events=@()
 while([DateTime]::UtcNow -lt $deadline){
  if(Test-Path $report){
   $f=[IO.File]::Open($report,'Open','Read',[IO.FileShare]::ReadWrite);$reader=[IO.StreamReader]::new($f,[Text.Encoding]::UTF8)
   try{$raw=$reader.ReadToEnd()}finally{$reader.Dispose()}
   $events=@($raw -split "`n"|Where-Object {$_.Trim().EndsWith('}')}|ForEach-Object {$_|ConvertFrom-Json})
   if(@($events|Where-Object {$_.event -eq 'exiting'}).Count -eq 1){break}
  }
  Start-Sleep -Milliseconds 100
 }
 $began=@($events|Where-Object {$_.event -eq 'started'});$ended=@($events|Where-Object {$_.event -eq 'exiting'})
 if($began.Count -ne 1 -or $began[0].cwd -ne $cwd -or @($began[0].args).Count -ne 1 -or $began[0].args[0] -ne 'recovery-marker' -or $ended.Count -ne 1 -or $ended[0].code -ne 7){throw 'Configured custom CLI arguments/cwd or independent exit changed'}
 $null=Product $globalArgs @('stop',$owned);$owned=$null
 $before=Product $globalArgs @('history');[IO.File]::WriteAllText((Join-Path $root 'history-before.json'),$before)
 $offline=@('--backend',(Join-Path $root 'absent-backend.exe'),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-recovery-native','--state-dir',(Join-Path $root 'state'))
 $null=Product $offline @('metadata','check')
 $backup=Join-Path $root '备份 #% copy.sqlite3'
 $null=Product $offline @('metadata','backup',$backup)
 $hash=(Get-FileHash $backup).Hash
 $null=Product $offline @('metadata','backup',$backup) $false
 if((Get-FileHash $backup).Hash -ne $hash){throw 'Existing backup changed'}
 $restoreArgs=@('--backend',(Join-Path $root 'absent-backend.exe'),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-recovery-native','--state-dir',(Join-Path $root 'restored'))
 $null=Product $restoreArgs @('metadata','restore',$backup)
 $restored=Product $restoreArgs @('history')
 if(($restored|ConvertFrom-Json|ConvertTo-Json -Depth 10 -Compress) -ne ($before|ConvertFrom-Json|ConvertTo-Json -Depth 10 -Compress)){throw 'Restored historical evidence changed'}
 $target=Join-Path $root 'restored/metadata.sqlite3';$targetHash=(Get-FileHash $target).Hash
 $null=Product $restoreArgs @('metadata','restore',$backup) $false
 if((Get-FileHash $target).Hash -ne $targetHash){throw 'Restore replaced existing database'}
 $bad=Join-Path $root 'damaged.sqlite3';[IO.File]::WriteAllText($bad,'synthetic broken database')
 $badHash=(Get-FileHash $bad).Hash
 $null=Product $offline @('metadata','check',$bad) $false
 if((Get-FileHash $bad).Hash -ne $badHash){throw 'Checking changed damaged database'}
 $after=Product $globalArgs @('list','--json')
 if($after.Trim() -ne '[]'){throw 'Test backend not empty'}
 $checks.Add(@{scenario='TOML custom command and cwd; offline check, exclusive backup/restore, damaged source preserved';session_id=$start.session_id;run_id=$start.run_id;result='passed'})
}catch{$failure=$_.Exception.Message}finally{
 if($owned){try{$null=Product $globalArgs @('stop',$owned)}catch{}}
 $commands|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'commands.json') -Encoding utf8
 @{utc=[DateTime]::UtcNow.ToString('o');os=[Environment]::OSVersion.VersionString;product_sha256=(Get-FileHash $product).Hash;checks=$checks.ToArray();failure=$failure;cleanup_owned_backend=($null -eq $owned)}|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'result.json') -Encoding utf8
}
if($failure){throw $failure}
'Native metadata recovery and generic configured CLI: passed.'
