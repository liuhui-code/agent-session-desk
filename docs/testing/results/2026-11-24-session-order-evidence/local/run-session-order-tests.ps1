$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-order-native-20261124'
if(Test-Path $root){throw 'Fresh native root required'};New-Item -ItemType Directory $root|Out-Null
$entries=ConvertFrom-Json @'
[{"file": "session-order-session.exe", "package": "session", "sha256": "2c72dfe4feb9046683d6f8f4613f3a72447b1f05deb73cb66aceb4cd46d4bafd"}, {"file": "session-order-store.exe", "package": "store", "sha256": "137eae404a7e7cbba9411fb84727dd2842e62039291cbd45856ec2723fdea49c"}, {"file": "session-order-manager.exe", "package": "manager", "sha256": "db4dafcb20786b460de7de514d8f7bfdea7bafba517a09bb41462b5d50f4deee"}, {"file": "session-order-ui.exe", "package": "ui", "sha256": "32bd9e393dfbca0ee2a19e6558afa0940b8158fa461378b2849121ef8d97d05e"}]
'@
$results=@()
foreach($entry in $entries){$binary=Join-Path $base $entry.file;if((Get-FileHash $binary).Hash.ToLowerInvariant() -cne $entry.sha256){throw 'Test binary mismatch'};$log=Join-Path $root ($entry.package+'.log');& $binary '-test.v' '-test.count=1' '-test.timeout=4m' *> $log;$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package=$entry.package;binary_sha256=(Get-FileHash $binary).Hash;exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $binary}).Count};$results+=,$r
@{utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString;results=$results}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native package failed or skipped'}
}
Get-Content (Join-Path $root 'outcome.json') -Raw -Encoding utf8
