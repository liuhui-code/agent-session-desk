$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'f6-reader-20261123';$output=Join-Path $root 'driver-green02.log'
$ErrorActionPreference='Continue'
& (Join-Path $base 'pwsh7/pwsh.exe') -NoLogo -NoProfile -File (Join-Path $root 'test-windows-resource-driver.ps1') -ConptyDirectory (Join-Path $root 'candidate-native') *> $output
$code=$LASTEXITCODE;$ErrorActionPreference='Stop';$text=[IO.File]::ReadAllText($output)
$text
@{phase='driver-green02';exit_code=$code;expected=0}|ConvertTo-Json
if($code -ne 0){throw 'Unexpected driver environment regression result'}
