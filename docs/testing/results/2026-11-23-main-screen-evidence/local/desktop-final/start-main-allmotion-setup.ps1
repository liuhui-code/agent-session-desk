$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$name='AsdMainAllMotionSetup20261123'
if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Setup task already exists'}
$code=@'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
try{& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'setup-main-allmotion-desktop.ps1'))));$result=@{result='passed';is_administrator=([Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator);session=(Get-Process -Id $PID).SessionId}}
catch{$result=@{result='failed';error=$_.Exception.Message;stack=$_.ScriptStackTrace}}
$result|ConvertTo-Json|Set-Content (Join-Path $base 'main-allmotion-setup-outcome.json') -Encoding utf8
'@
$action=New-ScheduledTaskAction -Execute powershell.exe -Argument ('-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code)))
$principal=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask $name -Action $action -Principal $principal -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 2))|Out-Null
Start-ScheduledTask $name
$file=Join-Path $base 'main-allmotion-setup-outcome.json';$until=[DateTime]::UtcNow.AddSeconds(20)
while(!(Test-Path $file) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if(!(Test-Path $file)){throw 'Setup not observed'}
Get-Content $file -Raw -Encoding utf8
$until=[DateTime]::UtcNow.AddSeconds(5);while((Get-ScheduledTask $name).State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if((Get-ScheduledTask $name).State -eq 'Running'){throw 'Setup still running'}
Unregister-ScheduledTask $name -Confirm:$false
$r=[IO.File]::ReadAllText($file)|ConvertFrom-Json;if($r.result -ne 'passed'){throw 'Setup failed'}
