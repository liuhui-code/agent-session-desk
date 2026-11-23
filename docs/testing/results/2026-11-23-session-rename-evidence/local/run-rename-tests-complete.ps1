$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'session-rename-20261123'
$test=Join-Path $base 'rename-ui-tests-complete.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '54c0f1eb621b4876964bfa1b5c9cf8e6347e99bdcc268af5832f314f7f65e746'){throw 'Final test differs'}
$log=Join-Path $root 'ui-tests-complete.log'
& $test '-test.v' '-test.count=1' *> $log
$code=$LASTEXITCODE
$text=[IO.File]::ReadAllText($log)
$result=@{package='ui';utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash}
$result|ConvertTo-Json|Set-Content (Join-Path $root 'ui-tests-complete-result.json') -Encoding utf8
$result|ConvertTo-Json
if($code -ne 0 -or $result.fail -or $result.skip){throw 'Final UI tests failed'}
Compress-Archive -LiteralPath $log,(Join-Path $root 'ui-tests-complete-result.json') -DestinationPath (Join-Path $base 'rename-complete-ui-tests-export.zip')
