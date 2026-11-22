$ErrorActionPreference='Stop'
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/layout-20261123'
@{label='registration';from=60;to=60}|ConvertTo-Json|Set-Content (Join-Path $root 'drag-request.json') -Encoding utf8
$ErrorActionPreference='Stop'
$name='AsdInjectedDrag3_20261123'
$previous=Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue;if($previous){if($previous.State -eq 'Running'){throw 'Paste task still running'};Unregister-ScheduledTask -TaskName $name -Confirm:$false}
$code='& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $env:USERPROFILE "asd-acceptance-20261122/injected-drag-run3.ps1"))))'
$a=New-ScheduledTaskAction -Execute powershell.exe -Argument ('-NoProfile -WindowStyle Hidden -STA -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code)))
$p=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
Register-ScheduledTask -TaskName $name -Action $a -Principal $p -Settings (New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 22))|Out-Null
Start-ScheduledTask $name
