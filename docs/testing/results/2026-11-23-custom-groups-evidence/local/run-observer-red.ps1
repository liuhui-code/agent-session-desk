$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'custom-groups-observer-red-20261123'
if(Test-Path $root){throw 'Fresh diagnostic root required'};New-Item -ItemType Directory $root|Out-Null
$test=Join-Path $base 'groups-store-observer-red.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '9e16332b04c78e0e69432544158f6fc93ef70c6979e91cb1a6ff932c2dcfc242'){throw 'Diagnostic executable differs'}
$log=Join-Path $root 'observer-red.log'
& $test '-test.v' '-test.run=^TestOutputObserverDoesNotPreventPublishedRecordRotation$' '-test.count=1' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
@{exit_code=$code;expected_red=$true;sharing_violation=$text.Contains('independent observer prevented pruning');binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count;utc=[DateTime]::UtcNow.ToString('o')}|ConvertTo-Json|Set-Content (Join-Path $root 'result.json') -Encoding utf8
Get-Content (Join-Path $root 'result.json') -Raw -Encoding utf8
if($code -eq 0 -or !$text.Contains('independent observer prevented pruning')){throw 'Expected independent-observer RED not observed'}
