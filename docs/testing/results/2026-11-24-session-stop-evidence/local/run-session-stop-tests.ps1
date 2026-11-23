$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-stop-native-20261124'
if(Test-Path $root){throw 'Fresh native root required'};New-Item -ItemType Directory $root|Out-Null
$entries=ConvertFrom-Json @'
[{"file": "session-stop-manager.exe", "package": "manager", "sha256": "452b41b96b7cbb2866a19ab262bc13721a31acf8e1d2c82fd73f1482b15162ee"}, {"file": "session-stop-ui.exe", "package": "ui", "sha256": "96263c323f1750a2e9ed0e1dd2d12f471d9a908c29056d1296a19b72b162616d"}]
'@
$results=@()
foreach($entry in $entries){$binary=Join-Path $base $entry.file;if((Get-FileHash $binary).Hash.ToLowerInvariant() -cne $entry.sha256){throw 'Test binary mismatch'};$log=Join-Path $root ($entry.package+'.log');& $binary '-test.v' '-test.count=1' '-test.timeout=4m' *> $log;$code=$LASTEXITCODE;$text=[IO.File]::ReadAllText($log)
$r=@{package=$entry.package;binary_sha256=(Get-FileHash $binary).Hash;exit_code=$code;top_level_pass=[regex]::Matches($text,'(?m)^--- PASS:').Count;all_pass=[regex]::Matches($text,'(?m)^\s*--- PASS:').Count;fail=[regex]::Matches($text,'(?m)^\s*--- FAIL:').Count;skip=[regex]::Matches($text,'(?m)^\s*--- SKIP:').Count;remaining_exact_image=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $binary}).Count};$results+=,$r
@{utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString;results=$results}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8
if($code -ne 0 -or $r.fail -or $r.skip){throw 'Native package failed or skipped'}
}
Get-Content (Join-Path $root 'outcome.json') -Raw -Encoding utf8
