$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'custom-groups-observer-green-20261123'
if(Test-Path $root){throw 'Fresh diagnostic root required'};New-Item -ItemType Directory $root|Out-Null
$test=Join-Path $base 'groups-store-observer-green.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '740ba3ce58b0a48d99b762ac9bd4541836a541a61d01c4e14e556c42aebfc0ba'){throw 'Diagnostic executable differs'}
$log=Join-Path $root 'observer-green.log'
& $test '-test.v' '-test.run=^TestOutput(PublicationNeverExposesIncompleteOrExcessFiles|ObserverDoesNotPreventPublishedRecordRotation)$' '-test.count=20' '-test.timeout=2m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
@{exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^--- FAIL:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count;utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json|Set-Content (Join-Path $root 'result.json') -Encoding utf8
Get-Content (Join-Path $root 'result.json') -Raw -Encoding utf8
if($code -ne 0){throw 'Independent observer regression failed'}
