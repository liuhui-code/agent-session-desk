$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$name='AsdBundlePermissionProbe20261123'
if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Task already exists'}
$code='& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $env:USERPROFILE "asd-acceptance-20261122/probe-bundle-permissions.ps1"))))'
$action=New-ScheduledTaskAction -Execute powershell.exe -Argument ('-NoProfile -NonInteractive -WindowStyle Hidden -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code)))
$principal=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask $name -Action $action -Principal $principal -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 2))|Out-Null
Start-ScheduledTask $name
Start-Sleep -Milliseconds 800
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/bundle-desktop-20261123';$file=Join-Path $root 'permission-probe-desktop.json';$until=[DateTime]::UtcNow.AddSeconds(15)
while(!(Test-Path $file) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if(!(Test-Path $file)){throw 'Permission probe not observed'}
$identity=[Security.Principal.WindowsIdentity]::GetCurrent();$principal=[Security.Principal.WindowsPrincipal]::new($identity)
@{is_administrator=$principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator);session=(Get-Process -Id $PID).SessionId}|ConvertTo-Json|Set-Content (Join-Path $root 'permission-probe-ssh.json') -Encoding utf8
Get-Content $file -Raw -Encoding utf8
$until=[DateTime]::UtcNow.AddSeconds(5);while((Get-ScheduledTask $name).State -eq 'Running' -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 100}
if((Get-ScheduledTask $name).State -eq 'Running'){throw 'Probe still running'}
Unregister-ScheduledTask $name -Confirm:$false
