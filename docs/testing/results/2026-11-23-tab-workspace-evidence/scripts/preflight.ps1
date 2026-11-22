$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$terminal=(Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1).FullName
@{utc=[DateTime]::UtcNow.ToString('o');terminal=$terminal;windows=[Environment]::OSVersion.VersionString;active=@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal}|Select-Object Id,SessionId,StartTime,MainWindowHandle);settings=(Join-Path ([IO.Path]::GetDirectoryName($terminal)) 'settings/settings.json');backend_sha256=(Get-FileHash (Join-Path $base 'backend/psmux.exe')).Hash;tasks=@(Get-ScheduledTask -TaskName '*Asd*' -ErrorAction SilentlyContinue|Select-Object TaskName,State)}|ConvertTo-Json -Depth 5
