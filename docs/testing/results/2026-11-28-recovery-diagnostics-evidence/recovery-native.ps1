$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$d=Join-Path $env:USERPROFILE 'asd-diagnostics-20261128';$p=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$hashes=Get-Content -Raw (Join-Path $d 'recovery-hashes.json')|ConvertFrom-Json
foreach($f in $hashes.PSObject.Properties){if((Get-FileHash (Join-Path $d $f.Name)).Hash.ToLowerInvariant()-cne $f.Value){throw 'Image hash guard'}}
$env:ASD_RECOVERY_NATIVE_BUNDLE=Join-Path $p 'terminal-runtime';$env:ASD_RECOVERY_PRODUCT=Join-Path $d 'native-product.exe'
$env:ASD_NATIVE_MESSAGING_BUNDLE=$env:ASD_RECOVERY_NATIVE_BUNDLE;$env:ASD_NATIVE_MESSAGING_FIXTURE=Join-Path $d 'asd-fixture.exe'
if((Get-FileHash (Join-Path $env:ASD_RECOVERY_NATIVE_BUNDLE 'psmux.exe')).Hash.ToLowerInvariant()-cne 'ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b'){throw 'Backend guard'}
$r=@();foreach($name in @('cli','ui')){$pattern='TestWindowsNativeRecovery(UIBridge|CLI)$';if($name-eq 'ui'){$pattern='Test(Recovery|Diagnostics|DiagnosticExport|WindowsNativeDiagnostics)'};$child=Start-Process (Join-Path $d ('final-'+$name+'.test.exe')) -ArgumentList '-test.run',$pattern,'-test.v','-test.timeout','180s' -NoNewWindow -PassThru -Wait -RedirectStandardOutput (Join-Path $d ('final-'+$name+'.log')) -RedirectStandardError (Join-Path $d ('final-'+$name+'-stderr.log'));$r+=@{package=$name;exit_code=$child.ExitCode}}
$r|ConvertTo-Json|Set-Content (Join-Path $d 'recovery-results.json');$r|ConvertTo-Json
