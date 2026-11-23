$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'footer-input-tests-final-20261123'
if(Test-Path $root){throw 'Fresh native root required'};New-Item -ItemType Directory $root|Out-Null
$test=Join-Path $base 'footer-input-ui.exe';if((Get-FileHash $test).Hash.ToLowerInvariant() -cne 'bf73af57a4307cd2c74ab090f6c6dfc799567e45b59105ed0ebadfc6af76db9a'){throw 'Test binary differs'}
$log=Join-Path $root 'ui-tests.log'; & $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package='ui';exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count}
@{utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString;results=@($r)}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8
Get-Content (Join-Path $root 'outcome.json') -Raw -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native regression failed'}
