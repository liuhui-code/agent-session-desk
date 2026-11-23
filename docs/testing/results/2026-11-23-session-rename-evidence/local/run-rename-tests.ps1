$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'session-rename-20261123'
if(Test-Path $root){throw 'Validation root already exists; inspect before continuation'}
New-Item -ItemType Directory -Path $root|Out-Null
$results=@()
$test=Join-Path $base 'rename-session-tests.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne 'fddcd520579d8b65690c31c12d7f3a07fbc79595b265140e9505ee886f1808b1'){throw 'Test EXE differs'}
$log=Join-Path $root 'session-tests.log'
& $test '-test.v' '-test.count=1' *> $log
$code=$LASTEXITCODE
$text=[IO.File]::ReadAllText($log)
$result=@{package='session';utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash}
$results+=$result
$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'native-test-results.json') -Encoding utf8
if($code -ne 0 -or $result.fail){$result|ConvertTo-Json;throw 'Native suite failed; inspect exact result before continuation'}
$test=Join-Path $base 'rename-store-tests.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '37fbc38ac946c1555bf718cb37dcb61548ceaad4d0e127c1ee91b7023085b3a5'){throw 'Test EXE differs'}
$log=Join-Path $root 'store-tests.log'
& $test '-test.v' '-test.count=1' *> $log
$code=$LASTEXITCODE
$text=[IO.File]::ReadAllText($log)
$result=@{package='store';utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash}
$results+=$result
$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'native-test-results.json') -Encoding utf8
if($code -ne 0 -or $result.fail){$result|ConvertTo-Json;throw 'Native suite failed; inspect exact result before continuation'}
$test=Join-Path $base 'rename-manager-tests.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '55823326fbc4d0934be38fc848f71a9473013d5061200a5e62060ef6651b5891'){throw 'Test EXE differs'}
$log=Join-Path $root 'manager-tests.log'
& $test '-test.v' '-test.count=1' *> $log
$code=$LASTEXITCODE
$text=[IO.File]::ReadAllText($log)
$result=@{package='manager';utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash}
$results+=$result
$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'native-test-results.json') -Encoding utf8
if($code -ne 0 -or $result.fail){$result|ConvertTo-Json;throw 'Native suite failed; inspect exact result before continuation'}
$test=Join-Path $base 'rename-ui-tests.exe'
if((Get-FileHash $test).Hash.ToLowerInvariant() -cne '17db85add4b1e5629f2be4f03aba657bd75cabb1793bd05115c631f0455382fe'){throw 'Test EXE differs'}
$log=Join-Path $root 'ui-tests.log'
& $test '-test.v' '-test.count=1' *> $log
$code=$LASTEXITCODE
$text=[IO.File]::ReadAllText($log)
$result=@{package='ui';utc=[DateTime]::UtcNow.ToString('o');exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;sha256=(Get-FileHash $test).Hash}
$results+=$result
$results|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'native-test-results.json') -Encoding utf8
if($code -ne 0 -or $result.fail){$result|ConvertTo-Json;throw 'Native suite failed; inspect exact result before continuation'}
$results|ConvertTo-Json -Depth 6
$files=@('native-test-results.json','session-tests.log','store-tests.log','manager-tests.log','ui-tests.log')
Compress-Archive -LiteralPath ($files|ForEach-Object {Join-Path $root $_}) -DestinationPath (Join-Path $base 'rename-model-tests-export.zip')
Get-FileHash (Join-Path $base 'rename-model-tests-export.zip')|ConvertTo-Json
