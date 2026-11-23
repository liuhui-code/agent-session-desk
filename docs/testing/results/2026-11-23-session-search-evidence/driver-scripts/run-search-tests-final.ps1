$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-search-20261123';$test=Join-Path $base 'search-ui-tests-final.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '980b1d5e10f9041e91be920fd9e3e38e562366020c74efe25ad50d12d00cf037'){throw 'Test EXE differs'}
& $test '-test.v' '-test.count=1' *> (Join-Path $root 'ui-tests-final.log');$code=$LASTEXITCODE
$text=[IO.File]::ReadAllText((Join-Path $root 'ui-tests-final.log'))
$result=@{utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;tests_sha256=(Get-FileHash $test).Hash;scope='Native Windows model/render tests with service fixtures; actual desktop flow recorded separately'}
$result|ConvertTo-Json|Set-Content (Join-Path $root 'ui-test-final-result.json') -Encoding utf8
$result|ConvertTo-Json
if($code -ne 0 -or $result.fail -or $result.skip){throw 'Native UI suite did not fully pass'}
