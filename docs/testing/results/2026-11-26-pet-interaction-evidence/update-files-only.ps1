$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.Encoding]::UTF8
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$out=Join-Path $base 'pet-fix-files-20261126';if(Test-Path $out){throw 'Fresh file handoff required'}
$path=Join-Path $package 'alpha.json';$manifest=Get-Content -Raw $path|ConvertFrom-Json
if($manifest.entrypoint -cne 'asd-integrated-74a86959.exe'){throw 'Daily entry changed'}
foreach($file in $manifest.files.PSObject.Properties){if((Get-FileHash (Join-Path $package $file.Name)).Hash.ToLowerInvariant() -cne $file.Value){throw 'Existing package changed'}}
$hash='1897bc9dc73fe519b256a073b368695d3e6accc3495933189ba9d68d9d167105'
$source=Join-Path $base 'asd-pet-fix.exe';$name='asd-pet-1897bc9d.exe';$target=Join-Path $package $name
if((Get-FileHash $source).Hash.ToLowerInvariant() -cne $hash -or (Test-Path $target)){throw 'Candidate hash/destination guard failed'}
$tests=Get-Content -Raw (Join-Path $base 'pet-fix-native-20261126/tests.json')|ConvertFrom-Json
$desktop=Get-Content -Raw (Join-Path $base 'pet-fix-desktop-20261126/result.json')|ConvertFrom-Json
if($tests.Count -ne 2 -or @($tests|Where-Object{$_.exit_code -ne 0 -or $_.top_fail -ne 0 -or $_.top_skip -ne 0}).Count -or $desktop.result -cne 'verified'){throw 'Required tests not passed'}
New-Item -ItemType Directory $out|Out-Null
Copy-Item $path (Join-Path $out 'alpha-before-private.json')
Copy-Item $source $target
if((Get-FileHash $target).Hash.ToLowerInvariant() -cne $hash){throw 'Candidate copy mismatch'}
$manifest.files|Add-Member -NotePropertyName $name -NotePropertyValue $hash
$manifest.entrypoint=$name;$manifest.ui_source_commit='fbd9bfbb351296c4856644f6af022507322a1ba2'
$temp=Join-Path $package 'alpha-pet-fix-files.tmp';if(Test-Path $temp){throw 'Manifest staging path exists'}
[IO.File]::WriteAllText($temp,($manifest|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($true));Move-Item $temp $path -Force
$readback=Get-Content -Raw $path|ConvertFrom-Json
if($readback.entrypoint -cne $name){throw 'Entry readback mismatch'}
foreach($file in $readback.files.PSObject.Properties){if((Get-FileHash (Join-Path $package $file.Name)).Hash.ToLowerInvariant() -cne $file.Value){throw 'Final package hash mismatch'}}
$r=@{result='entry-updated-existing-uis-preserved';entrypoint=$name;product_sha256=$hash;manifest_payload_count=@($readback.files.PSObject.Properties).Count;all_manifest_hashes_match=$true;processes_closed=0;new_launches=0;metadata_restore_attempts=0;metadata_migration_attempts=0;terminal_input_actions=0;registry_writes=0;clipboard_calls=0;session_content_exported=0;private_manifest_backup=$true}
$r|ConvertTo-Json|Set-Content (Join-Path $out 'result.json') -Encoding utf8
$r|ConvertTo-Json
