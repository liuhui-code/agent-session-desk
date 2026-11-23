$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'main-parity-desktop-20261123'
$t=Get-ScheduledTask 'AsdMainParityDesktop20261123';$i=Get-ScheduledTaskInfo $t
$errors=@();foreach($n in @('desktop/driver-error.json','desktop/clipboard-restored.json','flow-failure.json')){$p=Join-Path $root $n;if(Test-Path $p){$errors+=@{file=$n;data=[IO.File]::ReadAllText($p)}}}
@{utc=[DateTime]::UtcNow.ToString('o');task_state=[string]$t.State;last_task_result=$i.LastTaskResult;root_files=@(Get-ChildItem $root -Recurse -File|Select-Object -ExpandProperty Name);observations=$errors;processes=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq (Join-Path $base 'asd-main-parity.exe') -or $_.ExecutablePath -eq (Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe')}|Select-Object ProcessId,ExecutablePath)}|ConvertTo-Json -Depth 8
