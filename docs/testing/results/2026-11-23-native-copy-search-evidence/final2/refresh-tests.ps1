$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'copy-final-20261123'
$test=Join-Path $base 'copy-final2-ui-tests.exe';$terminalTest=Join-Path $base 'copy-final2-terminal-tests.exe'
& $test '-test.v' *> (Join-Path $root 'ui-tests-final2.log');if($LASTEXITCODE -ne 0){throw 'Final UI tests failed'}
& $terminalTest '-test.v' *> (Join-Path $root 'terminal-tests-final2.log');if($LASTEXITCODE -ne 0){throw 'Final terminal tests failed'}
$prompt=Join-Path (Join-Path $root 'work') 'prompt.ps1';$shell=Join-Path $base 'pwsh7/pwsh.exe';$identities=@()
$owned=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $shell -and $_.CommandLine.Contains($prompt)})
if($owned.Count -ne 2){throw 'Expected the same two live test shells before UI replacement'}
foreach($item in $owned){$p=Get-Process -Id $item.ProcessId;if($p.Path -ne $shell){throw 'Shell identity changed'};$identities+=,@{pid=$p.Id;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}}
$identities|ConvertTo-Json -Depth 4|Set-Content (Join-Path $root 'shells-before-ui-replacement.json') -Encoding utf8
$r=[IO.File]::ReadAllText((Join-Path $root 'receipt.json'))|ConvertFrom-Json
[IO.File]::Copy((Join-Path $root 'receipt.json'),(Join-Path $root 'receipt-before-resize-fix.json'))
$r.app_sha256=(Get-FileHash (Join-Path $base 'copy-final2-product.exe')).Hash;$r.test_sha256=(Get-FileHash $test).Hash;$r.terminal_test_sha256=(Get-FileHash $terminalTest).Hash
$r|Add-Member -NotePropertyName updated_utc -NotePropertyValue ([DateTime]::UtcNow.ToString('o'))
$r|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'receipt.json') -Encoding utf8
Get-Content (Join-Path $root 'receipt.json') -Raw -Encoding utf8
