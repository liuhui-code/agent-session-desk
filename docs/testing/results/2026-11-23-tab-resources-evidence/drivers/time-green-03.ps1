$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tab-resources-20261123'
$pwsh=Join-Path $base 'pwsh7/pwsh.exe';$output=Join-Path $root 'time-green-03.log'
$ErrorActionPreference='Continue'
& $pwsh -NoLogo -NoProfile -File (Join-Path $root 'test-windows-fixture-journal.ps1') *> $output
$code=$LASTEXITCODE
$ErrorActionPreference='Stop'
[IO.File]::ReadAllText($output)
@{phase='time-green-03';test_exit_code=$code;expected_exit_code=0;pwsh_sha256=(Get-FileHash $pwsh).Hash}|ConvertTo-Json
if($code -ne 0){throw 'Unexpected UTC regression result.'}
