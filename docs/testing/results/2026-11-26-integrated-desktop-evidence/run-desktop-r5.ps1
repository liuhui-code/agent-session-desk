$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$artifacts=Get-Content -Raw (Join-Path $base 'integrated-artifacts-r3.json')|ConvertFrom-Json
foreach($item in $artifacts.PSObject.Properties){if((Get-FileHash (Join-Path $base $item.Name)).Hash.ToLowerInvariant() -cne $item.Value){throw 'Artifacts changed'}}
$unit=Join-Path $base 'integrated-native-r5-20261126';if(Test-Path $unit){throw 'Fresh test output required'}
New-Item -ItemType Directory $unit|Out-Null
$results=@()
foreach($case in @(@{exe='document-integration.test.exe';pattern='^Test(Markdown|WindowsMarkdownAssociation)'},@{exe='fixture-integration.test.exe';pattern='^TestAttentionFixture'},@{exe='cli-integration.test.exe';pattern='^Test(DocumentApp|PetFlag)'})){
 & (Join-Path $base $case.exe) '-test.v' '-test.run' $case.pattern '-test.timeout' '45s' 2>&1|Set-Content (Join-Path $unit ($case.exe+'.stdout')) -Encoding utf8
 $code=$LASTEXITCODE;$log=@(Get-Content (Join-Path $unit ($case.exe+'.stdout')))
 $results+=@{exe=$case.exe;exit_code=$code;top_pass=@($log|Where-Object{$_ -match '^--- PASS:'}).Count;top_fail=@($log|Where-Object{$_ -match '^--- FAIL:'}).Count;top_skip=@($log|Where-Object{$_ -match '^--- SKIP:'}).Count}
 if($code -ne 0){$log|Write-Output;throw 'Native units failed before desktop'}
}
$results|ConvertTo-Json -Depth 4|Set-Content (Join-Path $unit 'tests.json') -Encoding utf8
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-task.ps1')))) -Driver (Join-Path $base 'desktop-flow-r5.ps1') -Task 'AsdIntegratedDesktopR5_20261126' -Receipt (Join-Path $base 'integrated-desktop-run5-20261126/result.json') -Seconds 200
