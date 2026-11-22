[CmdletBinding()]
param(
 [Parameter(Mandatory=$true)][string]$ProductBinary,
 [Parameter(Mandatory=$true)][string]$FixtureBinary,
 [Parameter(Mandatory=$true)][string]$PsmuxBinary,
 [Parameter(Mandatory=$true)][string]$ShellBinary,
 [Parameter(Mandatory=$true)][string]$EvidenceDirectory,
 [string]$StoreTestBinary=''
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
$PSNativeCommandUseErrorActionPreference=$false
if (-not $IsWindows){throw 'Native Windows only.'}
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$product=(Resolve-Path $ProductBinary).Path
$fixture=(Resolve-Path $FixtureBinary).Path
$psmux=(Resolve-Path $PsmuxBinary).Path
$shell=(Resolve-Path $ShellBinary).Path
if ((Get-FileHash $psmux -Algorithm SHA256).Hash.ToLowerInvariant() -ne 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'){throw 'Wrong pinned psmux artifact.'}
$root=[IO.Path]::GetFullPath($EvidenceDirectory)
if (Test-Path $root){throw 'Choose fresh evidence directory.'}
New-Item -ItemType Directory -Path $root | Out-Null
$globalArgs=@('--backend',$psmux,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-exit-native')
$commands=[Collections.Generic.List[object]]::new()
$checks=[Collections.Generic.List[object]]::new()
$owned=[Collections.Generic.List[string]]::new()
$failure=$null
$shellIdentity=$null
$shellIdentity=$null
function Write-Utf8([string]$Path,[string]$Text){[IO.File]::WriteAllText($Path,$Text,[Text.UTF8Encoding]::new($false))}
function Invoke-Product([string[]]$Arguments){
 $errPath=Join-Path $root ("stderr-$($commands.Count).txt")
 $text=(& $product @globalArgs @Arguments 2>$errPath|Out-String);$code=$LASTEXITCODE
 $commands.Add(@{args=$Arguments;stdout=$text;stderr=[IO.File]::ReadAllText($errPath);code=$code})
 if ($code -ne 0){throw "CLI failed: $text $([IO.File]::ReadAllText($errPath))"}
 return $text
}
function Read-Fixture([string]$Path){
 if (-not (Test-Path $Path)){return @()}
 $f=[IO.File]::Open($Path,'Open','Read',[IO.FileShare]::ReadWrite);$r=[IO.StreamReader]::new($f,[Text.Encoding]::UTF8)
 try {$text=$r.ReadToEnd()}finally{$r.Dispose()}
 return @($text -split "`n"|Where-Object { $_.EndsWith('}') }|ForEach-Object {$_|ConvertFrom-Json})
}
try {
 foreach ($exitCode in @(0,7)){
  $report=Join-Path $root "exit-$exitCode.jsonl"
  $s=Invoke-Product @('start','--name',"exit-$exitCode",'--env',"ASD_FIXTURE_REPORT_PATH=$report",'--env','ASD_FIXTURE_EXIT_AFTER_START=1','--env',"ASD_FIXTURE_EXIT_CODE=$exitCode",'--',$fixture,"exit-marker-$exitCode")|ConvertFrom-Json
  $owned.Add($s.session_id)
  $deadline=[DateTime]::UtcNow.AddSeconds(20);$found=$null
  while ([DateTime]::UtcNow -lt $deadline){
   $rows=@((Invoke-Product @('list','--json'))|ConvertFrom-Json)
   $row=@($rows|Where-Object {$_.session_id -eq $s.session_id})
   if ($row.Count -eq 1 -and $row[0].terminal_state -eq 'exited' -and $row[0].PSObject.Properties['pane_exit']){$found=$row[0];break}
   Start-Sleep -Milliseconds 100
  }
  $events=@(Read-Fixture $report);$ended=@($events|Where-Object {$_.event -eq 'exiting'})
  if ($null -eq $found -or $ended.Count -ne 1 -or $ended[0].code -ne $exitCode){throw 'Independent exit not observed.'}
  $e=$found.pane_exit
  $first=$e;if ($s.PSObject.Properties['pane_exit']){$first=$s.pane_exit}
  if ($e.source -ne 'psmux.pane_dead' -or $e.pane_id -notmatch '^%[0-9]+$' -or -not $e.observed_at -or -not $e.PSObject.Properties['exit_code'] -or $null -ne $e.exit_code){throw 'Unverified exit code or missing provenance.'}
  $null=Invoke-Product @('stop',$s.session_id);$owned.Remove($s.session_id)|Out-Null
  $history=@((Invoke-Product @('history'))|ConvertFrom-Json)
  $h=@($history|Where-Object {$_.session.run_id -eq $s.run_id})
  if ($h.Count -ne 1 -or $h[0].session.terminal_state -ne 'unknown' -or $h[0].pane_exit.observed_at -ne $first.observed_at -or $null -ne $h[0].pane_exit.exit_code -or -not $h[0].stop_confirmed_at){throw 'Persisted observation changed.'}
  $checks.Add(@{scenario="exit-$exitCode";session_id=$s.session_id;run_id=$s.run_id;fixture_exit=$ended[0].code;observation=$e;result='passed'})
 }
 $report=Join-Path $root 'shell-child.jsonl'
 $identityPath=Join-Path $root 'shell-process.json'
 $script="`$p=Get-Process -Id `$PID; @{pid=`$PID;filetime=`$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=`$p.Path}|ConvertTo-Json -Compress|Set-Content -Encoding utf8 '$identityPath'; & '$fixture' shell-child; Write-Output 'SHELL_RETAINED_AFTER_CHILD'; Start-Sleep -Seconds 300"
 $s=Invoke-Product @('start','--name','shell-retained','--env',"ASD_FIXTURE_REPORT_PATH=$report",'--env','ASD_FIXTURE_EXIT_AFTER_START=1','--env','ASD_FIXTURE_EXIT_CODE=7','--',$shell,'-NoLogo','-NoProfile','-Command',$script)|ConvertFrom-Json
 $owned.Add($s.session_id)
 $deadline=[DateTime]::UtcNow.AddSeconds(20);$ended=@()
 while ([DateTime]::UtcNow -lt $deadline){$ended=@(Read-Fixture $report|Where-Object {$_.event -eq 'exiting'});if ($ended.Count -eq 1){break};Start-Sleep -Milliseconds 100}
 $rows=@((Invoke-Product @('list','--json'))|ConvertFrom-Json);$row=@($rows|Where-Object {$_.session_id -eq $s.session_id})
 if ($ended.Count -ne 1 -or $ended[0].code -ne 7 -or $row.Count -ne 1 -or $row[0].terminal_state -ne 'available' -or $row[0].PSObject.Properties['pane_exit']){throw 'Shell liveness manufactured agent exit or readiness.'}
 $shellIdentity=Get-Content -Raw -LiteralPath $identityPath|ConvertFrom-Json
 $p=Get-Process -Id $shellIdentity.pid -ErrorAction Stop
 if ($p.Path -ne $shell -or $p.Path -ne $shellIdentity.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $shellIdentity.filetime){throw 'Shell process identity mismatch.'}
 $null=Invoke-Product @('stop',$s.session_id);$owned.Remove($s.session_id)|Out-Null
 $history=@((Invoke-Product @('history'))|ConvertFrom-Json);$h=@($history|Where-Object {$_.session.run_id -eq $s.run_id})
 if ($h.Count -ne 1 -or $h[0].PSObject.Properties['pane_exit']){throw 'Stop became pane exit evidence.'}
 $checks.Add(@{scenario='shell-retained';session_id=$s.session_id;run_id=$s.run_id;fixture_exit=7;result='passed'})
 if ($StoreTestBinary){
  $test=(Resolve-Path $StoreTestBinary).Path;$output=& $test '-test.v' 2>&1|Out-String;$code=$LASTEXITCODE
  Write-Utf8 (Join-Path $root 'store-tests.log') $output
  if ($code -ne 0 -or $output.Contains('SKIP')){throw 'Native store tests failed or skipped.'}
  $checks.Add(@{scenario='native-store-tests';sha256=(Get-FileHash $test -Algorithm SHA256).Hash.ToLowerInvariant();result='passed'})
 }
} catch {$failure=$_.Exception.Message}finally{
 foreach ($id in @($owned.ToArray())){try{$null=Invoke-Product @('stop',$id);$owned.Remove($id)|Out-Null}catch{$failure="cleanup failed: $_; $failure"}}
 try{$rows=@((Invoke-Product @('list','--json'))|ConvertFrom-Json);if ($rows.Count -ne 0){throw 'Owned namespace not empty.'}}catch{$failure="cleanup verification failed: $_; $failure"}
 if ($null -ne $shellIdentity){
  $deadline=[DateTime]::UtcNow.AddSeconds(5)
  do {$p=Get-Process -Id $shellIdentity.pid -ErrorAction SilentlyContinue;if ($null -eq $p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $shellIdentity.filetime){break};Start-Sleep -Milliseconds 100}while ([DateTime]::UtcNow -lt $deadline)
  if ($null -ne $p -and $p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $shellIdentity.filetime){$failure="Owned shell process still alive; $failure"}
 }
 Write-Utf8 (Join-Path $root 'commands.json') (ConvertTo-Json -InputObject $commands.ToArray() -Depth 12)
 Write-Utf8 (Join-Path $root 'result.json') (ConvertTo-Json -Depth 12 -InputObject ([ordered]@{utc=[DateTime]::UtcNow.ToString('o');os=[Environment]::OSVersion.VersionString;product_sha256=(Get-FileHash $product -Algorithm SHA256).Hash.ToLowerInvariant();psmux_sha256=(Get-FileHash $psmux -Algorithm SHA256).Hash.ToLowerInvariant();checks=$checks.ToArray();remaining_owned=$owned.Count;failure=$failure}))
}
if ($failure){throw $failure}
Write-Output 'Native pane exit evidence: passed.'
