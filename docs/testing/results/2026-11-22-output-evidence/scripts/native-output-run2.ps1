$ErrorActionPreference='Stop';$PSNativeCommandUseErrorActionPreference=$false
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'm1-output-run2'
if(Test-Path $root){throw 'Use fresh evidence directory'}
New-Item -ItemType Directory $root|Out-Null
$product=Join-Path $base 'asd-output.exe';$fixture=Join-Path $base 'asd-fixture.exe';$psmux=Join-Path $base 'backend/psmux.exe'
if((Get-FileHash $psmux).Hash.ToLowerInvariant() -ne 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'){throw 'Wrong psmux pin'}
$globalArgs=@('--backend',$psmux,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-output-native','--state-dir',(Join-Path $root 'state'))
$commands=[Collections.Generic.List[object]]::new();$checks=[Collections.Generic.List[object]]::new();$owned=$null;$failure=$null;$handle=$null;$savedRoot=$null;$outputRoot=Join-Path $root 'state/output-records'
function Product([string[]]$g,[string[]]$a,[bool]$success=$true){
 $errPath=Join-Path $root "stderr-$($commands.Count).txt"
 $text=(& $product @g @a 2>$errPath|Out-String);$code=$LASTEXITCODE
 $commands.Add(@{global=$g;args=$a;code=$code;stdout=$text;stderr=[IO.File]::ReadAllText($errPath)})
 if(($success -and $code -ne 0) -or (!$success -and $code -eq 0)){throw "Unexpected CLI outcome $code : $([IO.File]::ReadAllText($errPath))"}
 return $text
}
function Events([string]$path){
 if(!(Test-Path $path)){return @()}
 $f=[IO.File]::Open($path,'Open','Read',[IO.FileShare]::ReadWrite);$r=[IO.StreamReader]::new($f,[Text.Encoding]::UTF8)
 try{$text=$r.ReadToEnd()}finally{$r.Dispose()}
 return @($text -split "`n"|Where-Object {$_.EndsWith('}')}|ForEach-Object {$_|ConvertFrom-Json})
}
try{
 foreach($name in @('store','cli','manager')){
  $exe=Join-Path $base "$name-output-tests.exe"
  $text=(& $exe '-test.v' '-test.timeout=120s' 2>&1|Out-String);$code=$LASTEXITCODE
  [IO.File]::WriteAllText((Join-Path $root "$name-tests.log"),$text)
  if($code -ne 0 -or $text.Contains('SKIP')){throw "$name native tests failed or skipped"}
  $checks.Add(@{scenario="$name native tests";sha256=(Get-FileHash $exe).Hash;result='passed'})
 }
 $cwd=Join-Path $root '自定义 目录';New-Item -ItemType Directory $cwd|Out-Null
 $report=Join-Path $root 'fixture.jsonl';$config=Join-Path $root 'agents.toml'
 $toml="schema_version = 1`n[templates.custom]`nname = 'output fixture'`nexecutable = '$fixture'`nargs = ['output-marker']`ncwd = '$cwd'`n[templates.custom.env]`nASD_FIXTURE_REPORT_PATH = '$report'`n"
 [IO.File]::WriteAllText($config,$toml,[Text.UTF8Encoding]::new($false))
 $s=Product $globalArgs @('--config',$config,'start','--template','custom')|ConvertFrom-Json;$owned=$s.session_id
 $deadline=[DateTime]::UtcNow.AddSeconds(15);$started=@()
 while([DateTime]::UtcNow -lt $deadline){$started=@(Events $report|Where-Object {$_.event -eq 'started'});if($started.Count -eq 1){break};Start-Sleep -Milliseconds 100}
 if($started.Count -ne 1 -or $started[0].cwd -ne $cwd -or $started[0].args[0] -ne 'output-marker'){throw 'Custom configuration/cwd not observed'}
 $handle=Get-Process -Id $started[0].pid
 if($handle.Path -ne $fixture){throw 'Fixture path mismatch'}
 $null=$handle.Handle
 @{pid=$handle.Id;filetime=$handle.StartTime.ToUniversalTime().ToFileTimeUtc();path=$handle.Path}|ConvertTo-Json|Set-Content (Join-Path $root 'fixture-identity.json') -Encoding utf8
 $deadline=[DateTime]::UtcNow.AddSeconds(15)
 do{$p=Product $globalArgs @('preview',$s.session_id)|ConvertFrom-Json;if($p.snapshot.text -match 'CLI fixture'){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $deadline)
 if($p.snapshot.text -notmatch 'CLI fixture'){throw 'Fixture screen unavailable'}
 $captures=@()
 foreach($i in 1..3){$captures+=,(Product $globalArgs @('record',$s.session_id)|ConvertFrom-Json)}
 $rows=@((Product $globalArgs @('output',$s.run_id))|ConvertFrom-Json)
 if($rows.Count -ne 2 -or $rows[0].record_id -ne $captures[2].record_id -or $rows[1].record_id -ne $captures[1].record_id){throw 'Wrong retained records'}
 foreach($r in $rows){if($r.run_id -ne $s.run_id -or $r.session_id -ne $s.session_id -or $r.source -ne 'screen_snapshot' -or $r.complete -or $r.text -notmatch 'CLI fixture'){throw 'Output provenance/completeness mismatch'}}
 $checks.Add(@{scenario='Configured command/cwd and explicit three captures retain two';result='passed'})
 $savedRoot=Join-Path $root 'saved-output-records';Move-Item $outputRoot $savedRoot
 New-Item -ItemType Directory $outputRoot|Out-Null
 $sentinel=Join-Path $outputRoot 'unrelated.txt';[IO.File]::WriteAllText($sentinel,'synthetic do not delete')
 $hash=(Get-FileHash $sentinel).Hash
 $null=Product $globalArgs @('record',$s.session_id) $false
 if($handle.HasExited -or (Get-FileHash $sentinel).Hash -ne $hash){throw 'Record failure damaged fixture or unrelated file'}
 Move-Item $outputRoot (Join-Path $root 'rejected-unknown-directory')
 $external=Join-Path $root 'external-sentinel';New-Item -ItemType Directory $external|Out-Null
 Copy-Item (Join-Path $savedRoot '*') $external -Force
 Copy-Item (Join-Path $savedRoot '.format') $external -Force
 $sentinel=Join-Path $external 'outside.txt';[IO.File]::WriteAllText($sentinel,'outside synthetic sentinel');$hash=(Get-FileHash $sentinel).Hash
 New-Item -ItemType Junction -Path $outputRoot -Target $external|Out-Null
 $null=Product $globalArgs @('output',$s.run_id) $false
 if($commands[$commands.Count-1].stderr -notmatch 'output directory must not be a link'){throw 'Junction was not rejected at the directory boundary'}
 $null=Product $globalArgs @('record',$s.session_id) $false
 if($handle.HasExited -or (Get-FileHash $sentinel).Hash -ne $hash){throw 'Junction guard touched external data or process'}
 [IO.Directory]::Delete($outputRoot)
 Move-Item $savedRoot $outputRoot;$savedRoot=$null
 $checks.Add(@{scenario='Unrecognized directory and NTFS junction refused; fixture and sentinels unchanged';result='passed'})
 $null=Product $globalArgs @('stop',$s.session_id);$owned=$null
 if(!$handle.WaitForExit(10000)){throw 'Exact fixture process did not exit after stop'}
 $offline=@('--backend',(Join-Path $root 'absent-backend.exe'),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-output-native','--state-dir',(Join-Path $root 'state'))
 $after=@((Product $offline @('output',$s.run_id))|ConvertFrom-Json)
 if(($after|ConvertTo-Json -Depth 8 -Compress) -ne ($rows|ConvertTo-Json -Depth 8 -Compress)){throw 'Saved output changed after stop/offline read'}
 $null=Product $globalArgs @('record',$s.session_id) $false
 $empty=Product $globalArgs @('list','--json');if($empty.Trim() -ne '[]'){throw 'Backend not empty after stop'}
 $checks.Add(@{scenario='Offline read after exact fixture exit; ended Run cannot record/restart';result='passed'})
}catch{$failure=$_.Exception.Message}finally{
 if($savedRoot -and (Test-Path $savedRoot)){
  if(Test-Path $outputRoot){$item=Get-Item -LiteralPath $outputRoot -Force;if($item.Attributes -band [IO.FileAttributes]::ReparsePoint){[IO.Directory]::Delete($outputRoot)}else{Move-Item $outputRoot (Join-Path $root 'failed-directory-retained')}}
  Move-Item $savedRoot $outputRoot
 }
 if($owned){try{$null=Product $globalArgs @('stop',$owned);$owned=$null}catch{}}
 if($handle -and !$handle.HasExited){$null=$handle.WaitForExit(10000)}
 $commands|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'commands.json') -Encoding utf8
 @{utc=[DateTime]::UtcNow.ToString('o');os=[Environment]::OSVersion.VersionString;product_sha256=(Get-FileHash $product).Hash;checks=$checks.ToArray();failure=$failure;cleanup_owned_backend=($null -eq $owned);fixture_ended=($null -ne $handle -and $handle.HasExited)}|ConvertTo-Json -Depth 7|Set-Content (Join-Path $root 'result.json') -Encoding utf8
}
if($failure){throw $failure}
'Native configured CLI, bounded output records and offline read: passed.'
