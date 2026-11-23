$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-search-20261123';$test=Join-Path $base 'search-ui-tests-final.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '694bb225b2e77e30c02f0c8b59262eb7f7e85dc4b79bb5b81245d5f5b336caa8'){throw 'Test EXE differs'}
& $test '-test.v' '-test.count=1' *> (Join-Path $root 'ui-tests-complete.log');$code=$LASTEXITCODE
$text=[IO.File]::ReadAllText((Join-Path $root 'ui-tests-complete.log'))
$result=@{utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;tests_sha256=(Get-FileHash $test).Hash;scope='Native Windows model/render tests with service fixtures; actual desktop flow recorded separately'}
$result|ConvertTo-Json|Set-Content (Join-Path $root 'ui-test-complete-result.json') -Encoding utf8
$result|ConvertTo-Json
if($code -ne 0 -or $result.fail -or $result.skip){throw 'Native UI suite did not fully pass'}
