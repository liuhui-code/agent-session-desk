$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$d=Join-Path $env:USERPROFILE 'asd-update-20261128';$p=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$hashes=Get-Content -Raw (Join-Path $d 'native-hashes.json')|ConvertFrom-Json
foreach($f in $hashes.PSObject.Properties){if((Get-FileHash (Join-Path $d $f.Name)).Hash.ToLowerInvariant()-cne $f.Value){throw 'Native image changed'}}
$env:ASD_UPDATE_BUNDLE=Join-Path $p 'terminal-runtime';$env:ASD_UPDATE_PRODUCT=Join-Path $d 'asd-update-product.exe';$env:ASD_UPDATE_CANDIDATE=Join-Path $d 'asd-update-candidate.exe';$env:ASD_UPDATE_FIXTURE=Join-Path $d 'asd-fixture.exe'
if((Get-FileHash (Join-Path $env:ASD_UPDATE_BUNDLE 'psmux.exe')).Hash.ToLowerInvariant()-cne 'ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b'){throw 'Development backend guard'}
$cases=@(@{name='updater';pattern='.'},@{name='instance';pattern='^TestWithStopped'},@{name='cli';pattern='TestUpdateProbe|TestUpdateApply|TestWindowsNativePortableUpdate'})
$r=@{source='db63c5a0a99ef77f075798ebead9c31814341a89';started=[DateTime]::UtcNow.ToString('o');os_version=(Get-CimInstance Win32_OperatingSystem).Version;tests=@()}
foreach($case in $cases){$child=Start-Process -FilePath (Join-Path $d ($case.name+'.test.exe')) -ArgumentList '-test.run',$case.pattern,'-test.v','-test.timeout','180s' -NoNewWindow -PassThru -Wait -RedirectStandardOutput (Join-Path $d ($case.name+'.log')) -RedirectStandardError (Join-Path $d ($case.name+'-stderr.log'));$r.tests+=@{package=$case.name;exit_code=$child.ExitCode}}
$r.completed=[DateTime]::UtcNow.ToString('o');$r|ConvertTo-Json -Depth 5|Set-Content (Join-Path $d 'native-result.json');$r|ConvertTo-Json -Depth 5
