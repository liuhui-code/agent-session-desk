param([switch]$Worker)
$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$out=Join-Path $base 'blocking-pet-run2-20261126'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$hashes=Get-Content -Raw (Join-Path $base 'blocking-pet-artifact-fingerprints.json')|ConvertFrom-Json
foreach($p in $hashes.PSObject.Properties){if((Get-FileHash (Join-Path $base $p.Name)).Hash.ToLowerInvariant() -cne $p.Value){throw ('Hash mismatch: '+$p.Name)}}
$env:ASD_NATIVE_MESSAGING_BUNDLE=Join-Path $package 'terminal-runtime'
$env:ASD_NATIVE_REPLY_CLI=Join-Path $base 'asd-blocking-pet.exe'
$env:ASD_NATIVE_MESSAGING_FIXTURE=Join-Path $base 'asd-fixture-blocking-pet.exe'
if($Worker){
 $code=1
 try{
  & (Join-Path $base 'pet-blocking.test.exe') '-test.v' '-test.run' '^Test(Demo|LivePet|WindowsNativePet|WindowsLivePet)' '-test.timeout' '120s' 2>&1|Set-Content (Join-Path $out 'pet.stdout') -Encoding utf8
  $code=$LASTEXITCODE
 }catch{$_|Out-String|Set-Content (Join-Path $out 'worker-error.txt')}
 $log=@(Get-Content (Join-Path $out 'pet.stdout') -ErrorAction SilentlyContinue)
 @{exit_code=$code;top_pass=@($log|Where-Object{$_ -match '^--- PASS:'}).Count;top_fail=@($log|Where-Object{$_ -match '^--- FAIL:'}).Count;top_skip=@($log|Where-Object{$_ -match '^--- SKIP:'}).Count;clipboard_calls=0;native_window_messages='owned pet only';physical_mouse_evidence=$false;wt_foreground_evidence=$false}|ConvertTo-Json -Depth 4|Set-Content (Join-Path $out 'pet.json') -Encoding utf8
 return
}
if(Test-Path $out){throw 'Fresh evidence directory required'}
New-Item -ItemType Directory $out|Out-Null
$results=@()
foreach($case in @(@{exe='ui-pet-blocking.test.exe';pattern='^Test(Attention|Companion|WindowsNativeAttentionOriginalForm)'},@{exe='cli-pet-blocking.test.exe';pattern='^Test(Pet|Attention)'})){
 & (Join-Path $base $case.exe) '-test.v' '-test.run' $case.pattern '-test.timeout' '180s' 2>&1|Set-Content (Join-Path $out ($case.exe+'.stdout')) -Encoding utf8
 $code=$LASTEXITCODE;$log=@(Get-Content (Join-Path $out ($case.exe+'.stdout')))
 $results+=@{exe=$case.exe;exit_code=$code;top_pass=@($log|Where-Object{$_ -match '^--- PASS:'}).Count;top_fail=@($log|Where-Object{$_ -match '^--- FAIL:'}).Count;top_skip=@($log|Where-Object{$_ -match '^--- SKIP:'}).Count}
 if($code -ne 0){$results|ConvertTo-Json -Depth 4|Set-Content (Join-Path $out 'tests.json');$log|Write-Output;throw 'Native focused tests failed'}
}
$results|ConvertTo-Json -Depth 4|Set-Content (Join-Path $out 'tests.json') -Encoding utf8
$hostConfig=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
if((Get-FileHash $hostConfig.shell.path).Hash.ToLowerInvariant() -cne $hostConfig.shell.sha256){throw 'Pinned PowerShell changed'}
$workerFile=Join-Path $base 'native-blocking-pet-run2.ps1'
$code="& ([ScriptBlock]::Create([IO.File]::ReadAllText('"+$workerFile.Replace("'","''")+"'))) -Worker"
$arguments='-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code))
$task='AsdBlockingPet20261126Run2'
if(Get-ScheduledTask -TaskName $task -ErrorAction SilentlyContinue){throw 'Fresh task name required'}
$action=New-ScheduledTaskAction -Execute $hostConfig.shell.path -Argument $arguments
$principal=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
$settings=New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 3)
Register-ScheduledTask -TaskName $task -Action $action -Principal $principal -Settings $settings|Out-Null
try{
 Start-ScheduledTask -TaskName $task
 $until=[DateTime]::UtcNow.AddSeconds(160)
 while(!(Test-Path (Join-Path $out 'pet.json')) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 400}
 $t=Get-ScheduledTask -TaskName $task
 if($t.Actions.Arguments -cne $arguments){throw 'Task identity changed'}
 for($i=0;$i -lt 30 -and $t.State -eq 'Running';$i++){Start-Sleep -Milliseconds 200;$t=Get-ScheduledTask -TaskName $task}
 if($t.State -eq 'Running'){throw 'Worker still running; no force stop'}
 if(!(Test-Path (Join-Path $out 'pet.json'))){throw 'No worker result; start not replayed'}
 $pet=Get-Content -Raw (Join-Path $out 'pet.json')|ConvertFrom-Json
 $os=Get-CimInstance Win32_OperatingSystem
 $summary=@{tests=$results;pet=$pet;task_result=(Get-ScheduledTaskInfo -TaskName $task).LastTaskResult;os_version=$os.Version;os_build=$os.BuildNumber;personal_state_migrated=$false;new_agent_integration=$false;physical_desktop_input=$false;clipboard_calls=0}
 $summary|ConvertTo-Json -Depth 6|Set-Content (Join-Path $out 'summary.json') -Encoding utf8
 $summary|ConvertTo-Json -Depth 6
 if($pet.exit_code -ne 0 -or $pet.top_skip -gt 0){Get-Content (Join-Path $out 'pet.stdout');throw 'Pet native tests failed or skipped'}
}finally{
 $t=Get-ScheduledTask -TaskName $task -ErrorAction SilentlyContinue
 if($t -and $t.Actions.Arguments -ceq $arguments -and $t.State -ne 'Running'){Unregister-ScheduledTask -TaskName $task -Confirm:$false}
}
