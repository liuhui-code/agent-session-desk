$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$out=Join-Path $base 'instance-components-r3-20261126'
if(Test-Path $out){throw 'Fresh native output required'};New-Item -ItemType Directory $out|Out-Null
$m=Get-Content -Raw (Join-Path $base 'instance-artifacts-r3.json')|ConvertFrom-Json
foreach($file in $m.PSObject.Properties){if((Get-FileHash (Join-Path $base $file.Name)).Hash.ToLowerInvariant()-cne $file.Value){throw 'Native artifact changed'}}
$results=@()
foreach($case in @(@{exe='instance-r3.test.exe';pattern='^Test(ScopeKey|WindowsInstance)'},@{exe='pet-instance-r3.test.exe';pattern='^Test(WindowMarker|WindowsCaptureWindow)'},@{exe='cli-instance-r3.test.exe';pattern='^Test(TerminalHost|PortableOptions|PinnedLocal|WindowToken|PetFlag|DocumentApp)'})){
 & (Join-Path $base $case.exe) '-test.v' '-test.run' $case.pattern '-test.timeout' '45s' 2>&1|Set-Content (Join-Path $out ($case.exe+'.stdout')) -Encoding utf8
 $code=$LASTEXITCODE;$log=@(Get-Content (Join-Path $out ($case.exe+'.stdout')))
 $results+=@{exe=$case.exe;exit_code=$code;top_pass=@($log|Where-Object{$_ -match '^--- PASS:'}).Count;top_fail=@($log|Where-Object{$_ -match '^--- FAIL:'}).Count;top_skip=@($log|Where-Object{$_ -match '^--- SKIP:'}).Count}
 if($code-ne 0){break}
}
$results|ConvertTo-Json -Depth 4|Set-Content (Join-Path $out 'result.json') -Encoding utf8
