$ErrorActionPreference='Stop'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$m=Get-Content -Raw (Join-Path $package 'alpha.json')|ConvertFrom-Json
$old=Join-Path $package 'asd-focus-3654c0e.exe'
$items=@(Get-Process -Name 'asd-focus-3654c0e' -ErrorAction SilentlyContinue)
$exact=@($items|Where-Object{$_.Path -ieq $old})
$noPath=@($items|Where-Object{!$_.Path})
@{old_named_count=$items.Count;old_exact_count=$exact.Count;old_path_unavailable=$noPath.Count;old_session_ids=@($exact|ForEach-Object{$_.SessionId});entry_old=($m.entrypoint -ceq 'asd-focus-3654c0e.exe');candidate_exists=(Test-Path (Join-Path $package 'asd-integrated-74a86959.exe'));backup_exists=(Test-Path (Join-Path $package 'data/rollback-integrated-20261126'));install_task_exists=(@(Get-ScheduledTask -TaskName 'AsdIntegratedDailyInstall_20261126' -ErrorAction SilentlyContinue).Count -gt 0)}|ConvertTo-Json -Depth 3
