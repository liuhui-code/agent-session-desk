$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'density-20261123'
Copy-Item (Join-Path $root 'receipt.json') (Join-Path $root 'receipt-before-help.json')
Move-Item (Join-Path $root 'ui-tests.log') (Join-Path $root 'ui-tests-before-help.log')
$test=Join-Path $base 'density-ui-tests.exe';$product=Join-Path $base 'density-ui.exe'
& $test '-test.v' *> (Join-Path $root 'ui-tests.log');if($LASTEXITCODE -ne 0){throw 'Native final UI tests failed'}
$r=[IO.File]::ReadAllText((Join-Path $root 'receipt.json'))|ConvertFrom-Json
$r.utc=[DateTime]::UtcNow.ToString('o');$r.app_sha256=(Get-FileHash $product).Hash;$r.test_sha256=(Get-FileHash $test).Hash
$r|ConvertTo-Json|Set-Content (Join-Path $root 'receipt.json') -Encoding utf8
Get-Content (Join-Path $root 'receipt.json') -Raw -Encoding utf8
