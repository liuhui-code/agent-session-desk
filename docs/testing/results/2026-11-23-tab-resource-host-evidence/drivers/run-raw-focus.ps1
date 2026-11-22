$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'f6-reader-20261123';$oldRoot=Join-Path $base 'tab-resources-20261123';$repo=Join-Path $env:USERPROFILE 'asdr-1123'
if(Test-Path (Join-Path $root 'evidence/summary.json')){throw 'Probe already terminal; do not rerun'}
[IO.File]::Copy((Join-Path $root 'raw-focus-probe.ps1'),(Join-Path $repo 'tools/ci/raw-focus-probe.ps1'))
$ErrorActionPreference='Continue'
& (Join-Path $base 'pwsh7/pwsh.exe') -NoLogo -NoProfile -File (Join-Path $repo 'tools/ci/raw-focus-probe.ps1') -RawRecorderBinary (Join-Path $root 'raw-keys.exe') -PsmuxBinary (Join-Path $base 'backend/psmux.exe') -ProductBinary (Join-Path $oldRoot 'asd-resource.exe') -FixtureBinary (Join-Path $oldRoot 'asd-fixture-resource.exe') -EvidenceDirectory (Join-Path $root 'evidence') *> (Join-Path $root 'probe.log')
$code=$LASTEXITCODE;$ErrorActionPreference='Stop'
$summary=[IO.File]::ReadAllText((Join-Path $root 'evidence/summary.json'))|ConvertFrom-Json
$events=[IO.File]::ReadAllLines((Join-Path $root 'evidence/events.jsonl'))|ConvertFrom-Json
@{probe_exit_code=$code;result=$summary.result;failure=$summary.failure;binary_sha256=$summary.binary_sha256;stages=@($events|Where-Object {$_.event -like 'raw-*-observed' -or $_.event -eq 'product-f6-observed'}|ForEach-Object {[pscustomobject]@{event=$_.event;data=$_.data}})}|ConvertTo-Json -Depth 8
