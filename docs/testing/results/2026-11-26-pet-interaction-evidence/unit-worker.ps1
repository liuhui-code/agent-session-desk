$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$out=Join-Path $base 'pet-fix-native-20261126';if(Test-Path $out){throw 'Fresh native output required'};New-Item -ItemType Directory $out|Out-Null
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$manifest=Get-Content -Raw (Join-Path $base 'pet-fix-artifacts.json')|ConvertFrom-Json
foreach($item in $manifest.PSObject.Properties){if((Get-FileHash (Join-Path $base $item.Name)).Hash.ToLowerInvariant() -cne $item.Value){throw 'Artifact changed'}}
$env:ASD_NATIVE_PET_INTERACTIVE='1'
$env:ASD_NATIVE_MESSAGING_BUNDLE=Join-Path $package 'terminal-runtime'
$env:ASD_NATIVE_MESSAGING_FIXTURE=Join-Path $base 'asd-fixture-integrated.exe'
$env:ASD_NATIVE_REPLY_CLI=Join-Path $base 'asd-pet-fix.exe'
$results=@()
foreach($case in @(@{exe='pet-fix.test.exe';pattern='^Test(Demo|LivePet|IdlePet|Blink|WindowsNativePet|WindowsIdleBlink|WindowsLivePet)'},@{exe='ui-pet-fix.test.exe';pattern='^Test(Attention|Companion)'})){
 & (Join-Path $base $case.exe) '-test.v' '-test.run' $case.pattern '-test.timeout' '90s' 2>&1|Set-Content (Join-Path $out ($case.exe+'.stdout')) -Encoding utf8
 $code=$LASTEXITCODE;$log=@(Get-Content (Join-Path $out ($case.exe+'.stdout')))
 $results+=@{exe=$case.exe;exit_code=$code;top_pass=@($log|Where-Object{$_ -match '^--- PASS:'}).Count;top_fail=@($log|Where-Object{$_ -match '^--- FAIL:'}).Count;top_skip=@($log|Where-Object{$_ -match '^--- SKIP:'}).Count}
 if($code -ne 0){break}
}
$results|ConvertTo-Json -Depth 4|Set-Content (Join-Path $out 'tests.json') -Encoding utf8
