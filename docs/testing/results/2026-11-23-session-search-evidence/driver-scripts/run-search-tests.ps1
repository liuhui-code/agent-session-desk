$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-search-20261123';$test=Join-Path $base 'search-ui-tests.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne 'dbb77e1fe274e9a17963f42233164fe4157c8433ee4eb56b17b18027904b7590'){throw 'Test EXE differs'}
& $test '-test.v' '-test.count=1' *> (Join-Path $root 'ui-tests.log');$code=$LASTEXITCODE
$text=[IO.File]::ReadAllText((Join-Path $root 'ui-tests.log'))
$result=@{utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;tests_sha256=(Get-FileHash $test).Hash;scope='Native Windows model/render tests with service fixtures; actual desktop flow recorded separately'}
$result|ConvertTo-Json|Set-Content (Join-Path $root 'ui-test-result.json') -Encoding utf8
$result|ConvertTo-Json
if($code -ne 0 -or $result.fail -or $result.skip){throw 'Native UI suite did not fully pass'}
