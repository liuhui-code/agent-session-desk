$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'f6-reader-20261123';$oldRoot=Join-Path $base 'tab-resources-20261123';$repo=Join-Path $env:USERPROFILE 'asdr-1123'
if(Test-Path (Join-Path $root 'candidate-evidence/summary.json')){throw 'Probe already terminal; do not rerun'}
[IO.File]::Copy((Join-Path $root 'raw-focus-conpty.ps1'),(Join-Path $repo 'tools/ci/raw-focus-conpty.ps1'))
$candidate=Join-Path $root 'candidate-native'
$expected=@{'conpty.dll'='E971856F89748DC3B458125E9AFDD79B5E9574A7C937676116D3CBBD3B71F492';'OpenConsole.exe'='B17E48B07EF153B420F0A6E178BACD31FB4C3390E21AE746586213FC0719A775'}
$verified=@(foreach($file in $expected.Keys){$path=Join-Path $candidate $file;$sig=Get-AuthenticodeSignature $path;$hash=(Get-FileHash $path).Hash;if($hash -ne $expected[$file] -or $sig.Status -ne 'Valid' -or $sig.SignerCertificate.Subject -notmatch 'CN=Microsoft Corporation'){throw 'Candidate identity/signature differs'};[pscustomobject]@{path=$path;sha256=$hash;signature_status=[string]$sig.Status;signer=$sig.SignerCertificate.Subject;version=(Get-Item $path).VersionInfo.FileVersion}})
$verified|ConvertTo-Json -Depth 4|Set-Content (Join-Path $root 'candidate-verification.json') -Encoding utf8
$ErrorActionPreference='Continue'
& (Join-Path $base 'pwsh7/pwsh.exe') -NoLogo -NoProfile -File (Join-Path $repo 'tools/ci/raw-focus-conpty.ps1') -ConptyDirectory $candidate -RawRecorderBinary (Join-Path $root 'raw-keys.exe') -PsmuxBinary (Join-Path $base 'backend/psmux.exe') -ProductBinary (Join-Path $oldRoot 'asd-resource.exe') -FixtureBinary (Join-Path $oldRoot 'asd-fixture-resource.exe') -EvidenceDirectory (Join-Path $root 'candidate-evidence') *> (Join-Path $root 'candidate-probe.log')
$code=$LASTEXITCODE;$ErrorActionPreference='Stop'
$summary=[IO.File]::ReadAllText((Join-Path $root 'candidate-evidence/summary.json'))|ConvertFrom-Json
$events=[IO.File]::ReadAllLines((Join-Path $root 'candidate-evidence/events.jsonl'))|ConvertFrom-Json
@{probe_exit_code=$code;result=$summary.result;failure=$summary.failure;binary_sha256=$summary.binary_sha256;events=@($events|Where-Object {$_.event -like 'raw-*-observed' -or $_.event -eq 'product-f6-observed' -or $_.event -eq 'candidate-host-observed'}|ForEach-Object {[pscustomobject]@{event=$_.event;focus=$_.data.focus}})}|ConvertTo-Json -Depth 5
