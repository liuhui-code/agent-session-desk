$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'group-removal-tests-20261123'
if(Test-Path $root){throw 'Fresh native root required'};New-Item -ItemType Directory $root|Out-Null
$results=@()
$test=Join-Path $base 'group-removal-store.exe';if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '12927bd9260bb404030db4fc67cc4a209b4b39bb977c40d076c32ba399a9c894'){throw 'Test binary differs'}
$log=Join-Path $root 'store-tests.log'; & $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='store';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count}
$results+=,$r;$results|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native regression failed'}
$test=Join-Path $base 'group-removal-manager.exe';if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '9af14f5fb87b4df593c092af86a77338585531609ab0e2bcb429aa8b889ed624'){throw 'Test binary differs'}
$log=Join-Path $root 'manager-tests.log'; & $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='manager';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count}
$results+=,$r;$results|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native regression failed'}
$test=Join-Path $base 'group-removal-ui.exe';if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '48eb9ced8058b4132d562b879a493af33c0fee5f2377e0a47e8cd770bcddc5f9'){throw 'Test binary differs'}
$log=Join-Path $root 'ui-tests.log'; & $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='ui';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count}
$results+=,$r;$results|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native regression failed'}
@{utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString;results=$results}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8;Get-Content (Join-Path $root 'outcome.json') -Raw -Encoding utf8
