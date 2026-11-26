$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$m=Get-Content -Raw (Join-Path $package 'alpha.json')|ConvertFrom-Json
$target=Join-Path $package 'asd-integrated-74a86959.exe'
$expected='74a86959355ba37ea641d08c8664602fd8daa8e8313c1b621fffe4986fe31d99'
$payloads=@($m.files.PSObject.Properties)
$allHashes=$true
foreach($item in $payloads){if((Get-FileHash (Join-Path $package $item.Name)).Hash.ToLowerInvariant() -cne $item.Value){$allHashes=$false}}
$ps=@(Get-Process -Name 'asd-integrated-74a86959' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target})
$old=@(Get-Process -Name 'asd-focus-3654c0e' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq (Join-Path $package 'asd-focus-3654c0e.exe')})
$petCount=0
foreach($p in $ps){$petCount+=@([DeskFlow]::Find('ASD Blocking Reminders',$p.Id)).Count}
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$bindingCount=0
foreach($h in @([DeskFlow]::HostWindows($hostInfo.terminal.path))){$bindingCount+=[DeskFlow]::OwnerBindings($h)}
$check=(& $target metadata check (Join-Path $package 'data/state/metadata.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
if($LASTEXITCODE -ne 0){throw 'Current metadata structural check failed'}
$backup=(& $target metadata check (Join-Path $package 'data/rollback-integrated-20261126/metadata-v5.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
if($LASTEXITCODE -ne 0){throw 'Backup structural check failed'}
$names=@('AsdIntegratedDesktopProbe20261126','AsdIntegratedOwnedNotes_20261126','AsdIntegratedCleanupR1_20261126','AsdIntegratedDailyInstall_20261126','AsdIntegratedDailyInstallR2_20261126')
foreach($n in 1..5){$names+='AsdIntegratedDesktopR'+$n+'_20261126'}
$remaining=@($names|Where-Object{Get-ScheduledTask -TaskName $_ -ErrorAction SilentlyContinue})
$fixtures=@(Get-Process -Name 'asd-fixture-integrated' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq (Join-Path $base 'asd-fixture-integrated.exe')})
$managers=@(Get-Process -Name 'asd-integrated-r1','asd-integrated-r3' -ErrorAction SilentlyContinue|Where-Object{$_.Path -in @((Join-Path $base 'asd-integrated-r1.exe'),(Join-Path $base 'asd-integrated-r3.exe'))})
$r=@{result='audited';utc=[DateTime]::UtcNow.ToString('o');entrypoint_correct=($m.entrypoint -ceq 'asd-integrated-74a86959.exe');source_commit_correct=($m.ui_source_commit -ceq '845a5b72c37ec6bd3569248812c67084b9e2f614');product_sha256=(Get-FileHash $target).Hash.ToLowerInvariant();expected_product_match=((Get-FileHash $target).Hash.ToLowerInvariant() -ceq $expected);all_manifest_payload_hashes_match=$allHashes;manifest_payload_count=$payloads.Count;daily_manager_count=$ps.Count;old_daily_manager_count=$old.Count;daily_pet_count=$petCount;owner_binding_count=$bindingCount;current_schema=$check.schema_version;current_integrity=$check.integrity;backup_schema=$backup.schema_version;backup_integrity=$backup.integrity;temporary_task_remaining=$remaining.Count;private_fixture_remaining=$fixtures.Count;private_manager_remaining=$managers.Count;session_content_exported=0;clipboard_calls=0;registry_writes=0;ui_input_actions=0}
$r|ConvertTo-Json -Depth 4|Set-Content (Join-Path $base 'daily-integrated-final-audit.json') -Encoding utf8
