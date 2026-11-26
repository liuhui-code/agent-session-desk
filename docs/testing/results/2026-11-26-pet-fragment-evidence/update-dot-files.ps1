$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$out=Join-Path $base 'pet-dot-files-20261126';if(Test-Path $out){throw 'Fresh file handoff required'}
$path=Join-Path $package 'alpha.json';$manifest=Get-Content -Raw $path|ConvertFrom-Json
if($manifest.entrypoint -cne 'asd-pet-1897bc9d.exe'){throw 'Daily entry changed'}
foreach($file in $manifest.files.PSObject.Properties){if((Get-FileHash (Join-Path $package $file.Name)).Hash.ToLowerInvariant() -cne $file.Value){throw 'Existing package changed'}}
$hash='0201abe593f953e2f1efcedd0d0c88dbaf54cfad1bbf4923e4b4c11901404ff3'
$source=Join-Path $base 'asd-pet-dot.exe';$name='asd-pet-dot-0201abe5.exe';$target=Join-Path $package $name
if((Get-FileHash $source).Hash.ToLowerInvariant() -cne $hash -or (Test-Path $target)){throw 'Candidate hash/destination guard failed'}
$tests=Get-Content -Raw (Join-Path $base 'pet-dot-native-20261126/result.json')|ConvertFrom-Json
if($tests.exit_code -ne 0 -or $tests.top_pass -ne 6 -or $tests.top_fail -ne 0 -or $tests.top_skip -ne 0){throw 'Required native tests not passed'}
$check=(& $source metadata check (Join-Path $package 'data/state/metadata.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
if($LASTEXITCODE -ne 0 -or $check.schema_version -ne 9 -or $check.integrity -cne 'ok'){throw 'Existing metadata check failed; no migration'}
New-Item -ItemType Directory $out|Out-Null
Copy-Item $path (Join-Path $out 'alpha-before-private.json')
Copy-Item $source $target
if((Get-FileHash $target).Hash.ToLowerInvariant() -cne $hash){throw 'Candidate copy mismatch'}
$manifest.files|Add-Member -NotePropertyName $name -NotePropertyValue $hash
$manifest.entrypoint=$name;$manifest.ui_source_commit='71c3ff0ec338afaf50cb93d8472463d5e8421a00'
$temp=Join-Path $package 'alpha-pet-dot.tmp';if(Test-Path $temp){throw 'Manifest staging path exists'}
[IO.File]::WriteAllText($temp,($manifest|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($true));Move-Item $temp $path -Force
$readback=Get-Content -Raw $path|ConvertFrom-Json
if($readback.entrypoint -cne $name){throw 'Entry readback mismatch'}
foreach($file in $readback.files.PSObject.Properties){if((Get-FileHash (Join-Path $package $file.Name)).Hash.ToLowerInvariant() -cne $file.Value){throw 'Final package hash mismatch'}}
$prior=Join-Path $package 'asd-pet-1897bc9d.exe'
$priorUI=@(Get-Process -Name 'asd-pet-1897bc9d' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $prior})
$newUI=@(Get-Process -Name 'asd-pet-dot-0201abe5' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target})
$r=@{result='entry-updated-existing-uis-preserved';entrypoint=$name;product_sha256=$hash;source_commit=$readback.ui_source_commit;manifest_payload_count=@($readback.files.PSObject.Properties).Count;all_manifest_hashes_match=$true;processes_closed=0;new_launches=0;metadata_restore_attempts=0;metadata_migration_attempts=0;schema_version=$check.schema_version;integrity=$check.integrity;terminal_input_actions=0;registry_writes=0;clipboard_calls=0;session_content_exported=0;private_manifest_backup=$true;prior_entry_ui_count=$priorUI.Count;new_entry_ui_count=$newUI.Count;native_task_absent=(!(Get-ScheduledTask -TaskName 'AsdPetDotUnits20261126' -ErrorAction SilentlyContinue));native_test_processes=@(Get-Process -Name 'pet-dot.test' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq (Join-Path $base 'pet-dot.test.exe')}).Count}
$r|ConvertTo-Json|Set-Content (Join-Path $out 'result.json') -Encoding utf8
$r|ConvertTo-Json
