$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$out=Join-Path $base 'markdown-run1-20261126'
if(Test-Path $out){throw 'Fresh evidence directory required'}
New-Item -ItemType Directory $out|Out-Null
$hashes=Get-Content -Raw (Join-Path $base 'markdown-artifact-fingerprints.json')|ConvertFrom-Json
foreach($item in $hashes.PSObject.Properties){if((Get-FileHash (Join-Path $base $item.Name)).Hash.ToLowerInvariant() -cne $item.Value){throw ('Artifact mismatch: '+$item.Name)}}
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$env:ASD_PSMUX_TEST_BINARY=Join-Path $package 'terminal-runtime/psmux.exe'
$env:ASD_PSMUX_TEST_SHA256='ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b'
$env:ASD_FIXTURE_TEST_BINARY=Join-Path $base 'asd-fixture-blocking-pet.exe'
if((Get-FileHash $env:ASD_FIXTURE_TEST_BINARY).Hash.ToLowerInvariant() -cne '6ed5ed77e84835dddf699b1ea9e695d04c0f489e403eb85ad3111e104c73388f'){throw 'Reused fixture fingerprint changed'}
$results=@()
foreach($case in @(@{name='document';pattern='^TestMarkdown'},@{name='backend';pattern='^Test(Markdown|WindowsPsmuxNativeMarkdown)'},@{name='manager';pattern='^TestMarkdown'},@{name='ui';pattern='^TestMarkdown'})){
 $exe=Join-Path $base ($case.name+'-md.test.exe');$log=Join-Path $out ($case.name+'.stdout')
 & $exe '-test.v' '-test.run' $case.pattern '-test.timeout' '120s' 2>&1|Set-Content $log -Encoding utf8
 $code=$LASTEXITCODE;$lines=@(Get-Content $log)
 $results+=@{name=$case.name;exit_code=$code;top_pass=@($lines|Where-Object{$_ -match '^--- PASS:'}).Count;top_fail=@($lines|Where-Object{$_ -match '^--- FAIL:'}).Count;top_skip=@($lines|Where-Object{$_ -match '^--- SKIP:'}).Count;sha256=(Get-FileHash $exe).Hash.ToLowerInvariant()}
 if($code -ne 0){$lines|Write-Output;break}
}
$os=Get-CimInstance Win32_OperatingSystem
$images=@('document-md.test.exe','backend-md.test.exe','manager-md.test.exe','ui-md.test.exe')
$owned=@(Get-CimInstance Win32_Process|Where-Object{$_.Name -in $images -and $_.ExecutablePath -ieq (Join-Path $base $_.Name)}|ForEach-Object{$_.ProcessId})
$summary=@{tests=$results;os_version=$os.Version;os_build=$os.BuildNumber;psmux_sha256=(Get-FileHash $env:ASD_PSMUX_TEST_BINARY).Hash.ToLowerInvariant();fixture_sha256=(Get-FileHash $env:ASD_FIXTURE_TEST_BINARY).Hash.ToLowerInvariant();personal_state_migrated=$false;daily_entry_modified=$false;default_viewer_launched=$false;shell_execute_observer=$true;physical_desktop_input=$false;clipboard_calls=0;remaining_test_images=$owned;backend_global_audit=$false}
$summary|ConvertTo-Json -Depth 6|Set-Content (Join-Path $out 'summary.json') -Encoding utf8
$summary|ConvertTo-Json -Depth 6
if(@($results|Where-Object{$_.exit_code -ne 0 -or $_.top_skip -ne 0}).Count -gt 0 -or $results.Count -ne 4){exit 1}
