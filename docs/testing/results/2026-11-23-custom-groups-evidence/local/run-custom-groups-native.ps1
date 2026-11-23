$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'custom-groups-model-run3-20261123'
if(Test-Path $root){throw 'Fresh model root required; inspect existing run before continuation'}
New-Item -ItemType Directory $root|Out-Null
$results=@();$paths=@()
$test=Join-Path $base 'groups-session-tests.exe'
$paths+=,$test
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '7b8865a12d7e6226741efa6183c01c70d22010914f89608cd3b9457e453ed693'){throw 'Test executable differs'}
$log=Join-Path $root 'session-tests.log'
& $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='session';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash;utc=[DateTime]::UtcNow.ToString('o')}
$results+=,$r;$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){$r|ConvertTo-Json;throw 'Native model suite failed; inspect result before continuation'}
$test=Join-Path $base 'groups-store-tests.exe'
$paths+=,$test
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '740ba3ce58b0a48d99b762ac9bd4541836a541a61d01c4e14e556c42aebfc0ba'){throw 'Test executable differs'}
$log=Join-Path $root 'store-tests.log'
& $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='store';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash;utc=[DateTime]::UtcNow.ToString('o')}
$results+=,$r;$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){$r|ConvertTo-Json;throw 'Native model suite failed; inspect result before continuation'}
$test=Join-Path $base 'groups-manager-tests.exe'
$paths+=,$test
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '1fdc7a01701150efa7c7a10907ec057f6e00538b77e7f9f90bf28d53f180cb0d'){throw 'Test executable differs'}
$log=Join-Path $root 'manager-tests.log'
& $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='manager';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash;utc=[DateTime]::UtcNow.ToString('o')}
$results+=,$r;$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){$r|ConvertTo-Json;throw 'Native model suite failed; inspect result before continuation'}
$test=Join-Path $base 'groups-config-tests.exe'
$paths+=,$test
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '302b691f26698ce660bbc3a7a2370eb3b590a7daaa676633b4af6d18f6b3e9fc'){throw 'Test executable differs'}
$log=Join-Path $root 'config-tests.log'
& $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='config';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash;utc=[DateTime]::UtcNow.ToString('o')}
$results+=,$r;$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){$r|ConvertTo-Json;throw 'Native model suite failed; inspect result before continuation'}
$test=Join-Path $base 'groups-ui-tests.exe'
$paths+=,$test
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne 'd6122b476d29d3c0f2c0ee2981262d717198d0632e57a199e53c20ddedc428c3'){throw 'Test executable differs'}
$log=Join-Path $root 'ui-tests.log'
& $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='ui';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash;utc=[DateTime]::UtcNow.ToString('o')}
$results+=,$r;$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){$r|ConvertTo-Json;throw 'Native model suite failed; inspect result before continuation'}
$remaining=@(Get-CimInstance Win32_Process|Where-Object{$paths -contains $_.ExecutablePath})
if($remaining.Count -ne 0){throw 'Test image still live'}
@{windows=[Environment]::OSVersion.VersionString;process_arch=[Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture.ToString();utc=[DateTime]::UtcNow.ToString('o');remaining_test_images=$remaining.Count;test_type='native Windows model/store/config tests; no desktop or target agent claim'}|ConvertTo-Json|Set-Content (Join-Path $root 'environment.json') -Encoding utf8
$files=@('results.json','environment.json','session-tests.log','store-tests.log','manager-tests.log','config-tests.log','ui-tests.log')
Compress-Archive -LiteralPath ($files|ForEach-Object{Join-Path $root $_}) -DestinationPath (Join-Path $base 'custom-groups-model-run3-export.zip')
$results|ConvertTo-Json -Depth 6
Get-FileHash (Join-Path $base 'custom-groups-model-run3-export.zip')|Select-Object Hash|ConvertTo-Json
