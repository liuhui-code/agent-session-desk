$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$m=Get-Content -Raw (Join-Path $package 'alpha.json')|ConvertFrom-Json
$hashes=$true;foreach($p in $m.files.PSObject.Properties){if((Get-FileHash (Join-Path $package $p.Name)).Hash.ToLowerInvariant() -cne $p.Value){$hashes=$false}}
$target=Join-Path $package 'asd-pet-1897bc9d.exe';$old=Join-Path $package 'asd-integrated-74a86959.exe'
$check=(& $target metadata check (Join-Path $package 'data/state/metadata.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
if($LASTEXITCODE -ne 0){throw 'Metadata structural check failed'}
$names=@('AsdPetUserProbe20261126','AsdPetFixUnits20261126','AsdPetFixDesktop20261126','AsdPetFixDaily20261126','AsdPetFixDailyR220261126','AsdPetFixRoles20261126')
$remaining=@($names|Where-Object{Get-ScheduledTask -TaskName $_ -ErrorAction SilentlyContinue})
$r=@{result='file-handoff-audited';entrypoint_correct=($m.entrypoint -ceq 'asd-pet-1897bc9d.exe');source_commit_correct=($m.ui_source_commit -ceq 'fbd9bfbb351296c4856644f6af022507322a1ba2');all_manifest_hashes_match=$hashes;manifest_payload_count=@($m.files.PSObject.Properties).Count;current_schema=$check.schema_version;current_integrity=$check.integrity;old_ui_count=@(Get-Process -Name 'asd-integrated-74a86959' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $old}).Count;new_ui_count=@(Get-Process -Name 'asd-pet-1897bc9d' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target}).Count;temporary_task_remaining=$remaining.Count;private_test_images_remaining=@(Get-Process -Name 'pet-fix.test','ui-pet-fix.test','asd-pet-fix','asd-fixture-integrated' -ErrorAction SilentlyContinue|Where-Object{$_.Path.StartsWith($base,[StringComparison]::OrdinalIgnoreCase)}).Count;ui_actions=0;clipboard_calls=0;registry_writes=0;session_content_exported=0}
$r|ConvertTo-Json
