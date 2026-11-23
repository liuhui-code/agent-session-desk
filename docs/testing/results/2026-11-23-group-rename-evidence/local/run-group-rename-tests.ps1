$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'group-rename-tests-20261123'
if(Test-Path $root){throw 'Fresh native root required'};New-Item -ItemType Directory $root|Out-Null
$results=@()
$test=Join-Path $base 'group-rename-store.exe';if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '94997e0c3e1943d29d818ae472b45271dedbf033dc1fc9114267ed98cf3dd34c'){throw 'Test binary differs'}
$log=Join-Path $root 'store-tests.log'; & $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='store';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count}
$results+=,$r;$results|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native regression failed'}
$test=Join-Path $base 'group-rename-manager.exe';if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '99ef8a42b95a5b617f6d16dc82b8c5f3f28db8b5b16af5761356dd7f94e03bdd'){throw 'Test binary differs'}
$log=Join-Path $root 'manager-tests.log'; & $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='manager';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count}
$results+=,$r;$results|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native regression failed'}
$test=Join-Path $base 'group-rename-ui.exe';if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '9f292aa3e3c88adc0dded58afe51b30d508c39e34037cdd68020c9da25db477f'){throw 'Test binary differs'}
$log=Join-Path $root 'ui-tests.log'; & $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='ui';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count}
$results+=,$r;$results|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'results.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native regression failed'}
@{utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString;results=$results}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8;Get-Content (Join-Path $root 'outcome.json') -Raw -Encoding utf8
