param([string]$Driver,[string]$Task,[string]$Receipt,[int]$Seconds=180)
$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
if((Get-FileHash $hostInfo.shell.path).Hash.ToLowerInvariant() -cne $hostInfo.shell.sha256){throw 'Pinned PowerShell changed'}
if((Test-Path $Receipt) -or (Get-ScheduledTask -TaskName $Task -ErrorAction SilentlyContinue)){throw 'Fresh task and receipt required'}
$tokens=$null;$parseErrors=$null;$null=[Management.Automation.Language.Parser]::ParseFile($Driver,[ref]$tokens,[ref]$parseErrors)
if($parseErrors.Count){throw 'Driver parsing failed before launch'}
$code="& ([ScriptBlock]::Create([IO.File]::ReadAllText('"+$Driver.Replace("'","''")+"')))"
$arguments='-NoProfile -NonInteractive -WindowStyle Hidden -STA -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code))
$action=New-ScheduledTaskAction -Execute $hostInfo.shell.path -Argument $arguments
$principal=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask -TaskName $Task -Action $action -Principal $principal -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 5))|Out-Null
try{
 Start-ScheduledTask -TaskName $Task
 $deadline=[DateTime]::UtcNow.AddSeconds($Seconds)
 while(!(Test-Path $Receipt) -and [DateTime]::UtcNow -lt $deadline){Start-Sleep -Milliseconds 300}
 $t=Get-ScheduledTask -TaskName $Task
 if($t.Actions.Arguments -cne $arguments){throw 'Task identity changed'}
 for($i=0;$i -lt 40 -and $t.State -eq 'Running';$i++){Start-Sleep -Milliseconds 200;$t=Get-ScheduledTask -TaskName $Task}
 if($t.State -eq 'Running'){throw 'Worker still running; no force stop'}
 if(!(Test-Path $Receipt)){throw 'No worker receipt; launch not repeated'}
 Get-Content -Raw $Receipt
}finally{
 $t=Get-ScheduledTask -TaskName $Task -ErrorAction SilentlyContinue
 if($t -and $t.Actions.Arguments -ceq $arguments -and $t.State -ne 'Running'){Unregister-ScheduledTask -TaskName $Task -Confirm:$false}
}
