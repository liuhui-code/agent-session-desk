$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new();$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$out="$b\detach-diagnosis"
$task=Get-ScheduledTask -TaskName AsdDetachDiagnosis20261122 -ErrorAction SilentlyContinue
if($task){if($task.State-eq'Running'){throw 'Diagnostic driver still running'};Unregister-ScheduledTask -TaskName $task.TaskName -Confirm:$false}
$terminal=(Get-ChildItem "$b\terminal" -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1).FullName
Copy-Item "$out\terminal-settings-before.json" (Join-Path (Split-Path $terminal) 'settings/settings.json')
$owned=@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object Path -eq $terminal|Select-Object Id,SessionId)
@{utc=[DateTime]::UtcNow.ToString('o');task_removed=$true;private_terminal_remaining=$owned;private_terminal_settings_restored=$true}|ConvertTo-Json -Depth 4|Set-Content "$out\task-cleanup.json" -Encoding UTF8
Compress-Archive -Path $out -DestinationPath "$b\detach-evidence.zip"
Get-Content "$out\task-cleanup.json" -Raw
