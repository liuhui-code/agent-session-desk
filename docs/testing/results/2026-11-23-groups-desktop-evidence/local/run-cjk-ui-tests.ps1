$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'groups-cjk-ui-20261123'
if(Test-Path $root){throw 'Fresh UI test root required'};New-Item -ItemType Directory $root|Out-Null
$test=Join-Path $base 'groups-cjk-ui-tests.exe';if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '055845ebdb4bfb8e07077edf74d805c44e863d07baf23f93595cbe5b6b39ebff'){throw 'UI executable differs'}
$log=Join-Path $root 'ui-tests.log'; & $test '-test.v' '-test.count=1' '-test.timeout=3m' *> $log
$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;binary_sha256=(Get-FileHash $test).Hash;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $test}).Count;utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString}
$r|ConvertTo-Json|Set-Content (Join-Path $root 'result.json') -Encoding utf8;$r|ConvertTo-Json
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native UI regression failed'}
