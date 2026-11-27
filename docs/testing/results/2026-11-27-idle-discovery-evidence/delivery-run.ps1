$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$d=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$legacy=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$p=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e';$hostInfo=Get-Content -Raw (Join-Path $p 'local-host.json')|ConvertFrom-Json
$task='AsdIdleDelivery1d842c0'
if(Get-ScheduledTask -TaskName $task -ErrorAction SilentlyContinue){throw 'Fresh task required'}
$settings=@{d=$d;legacy=$legacy;p=$p}|ConvertTo-Json -Compress
$cfg=[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($settings))
$bootstrap='$ErrorActionPreference="Stop";$ProgressPreference="SilentlyContinue";$c=([Text.Encoding]::UTF8.GetString([Convert]::FromBase64String("'+$cfg+'"))|ConvertFrom-Json);$d=$c.d;[IO.File]::WriteAllText((Join-Path $d "driver-start.txt"),[DateTime]::UtcNow.ToString("o"));try{& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $d "idle-delivery.ps1"))))}catch{@{result="bootstrap-failed";error=$_.Exception.Message;stack=$_.ScriptStackTrace}|ConvertTo-Json|Set-Content (Join-Path $d "result.json")}'
$taskArguments='-NoProfile -NonInteractive -WindowStyle Hidden -STA -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($bootstrap))
$action=New-ScheduledTaskAction -Execute $hostInfo.shell.path -Argument $taskArguments
$principal=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask -TaskName $task -Action $action -Principal $principal -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 30))|Out-Null
$started=[DateTime]::Now
Start-ScheduledTask -TaskName $task
try{
 $end=[DateTime]::UtcNow.AddSeconds(300)
 do{
  $t=Get-ScheduledTask -TaskName $task
  if(Test-Path (Join-Path $d 'idle-discovery-delivery-1d842c0/result.json')){break}
  if([DateTime]::Now.Subtract($started).TotalSeconds-gt 8-and $t.State-ne 'Running'){
   $i=Get-ScheduledTaskInfo -TaskName $task
   if($i.LastRunTime-ge $started.AddSeconds(-1)){break}
  }
  Start-Sleep -Milliseconds 300
 }while([DateTime]::UtcNow-lt $end)
 for($n=0;$n-lt 30-and (Get-ScheduledTask -TaskName $task).State-eq 'Running';$n++){Start-Sleep -Milliseconds 200};$t=Get-ScheduledTask -TaskName $task
 $i=Get-ScheduledTaskInfo -TaskName $task
 @{state=[string]$t.State;last_task_result=$i.LastTaskResult;worker_started=(Test-Path (Join-Path $d 'driver-start.txt'));receipt=(Test-Path (Join-Path $d 'idle-discovery-delivery-1d842c0/result.json'))}|ConvertTo-Json|Set-Content (Join-Path $d 'idle-delivery-controller.json')
 if(Test-Path (Join-Path $d 'idle-discovery-delivery-1d842c0/result.json')){Get-Content -Raw (Join-Path $d 'idle-discovery-delivery-1d842c0/result.json')}
 Get-Content -Raw (Join-Path $d 'idle-delivery-controller.json')
}finally{
 $t=Get-ScheduledTask -TaskName $task -ErrorAction SilentlyContinue
 if($t-and $t.State-ne 'Running'-and $t.Actions.Arguments-ceq $taskArguments){Unregister-ScheduledTask -TaskName $task -Confirm:$false}
}
