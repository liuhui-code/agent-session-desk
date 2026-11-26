$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$out=Join-Path $base 'daily-integrated-install-20261126'
if(Test-Path $out){throw 'Fresh install receipt directory required'}
New-Item -ItemType Directory $out|Out-Null
$receipt=Join-Path $out 'result.json'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$manifestPath=Join-Path $package 'alpha.json';$manifest=Get-Content -Raw $manifestPath|ConvertFrom-Json
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$old=Join-Path $package $manifest.entrypoint
$oldHash='2aac7ca51c497a8943fa0b342d289a4f159cb33e5a022c0db4a97eedc292291f'
$expected='74a86959355ba37ea641d08c8664602fd8daa8e8313c1b621fffe4986fe31d99'
$name='asd-integrated-74a86959.exe';$target=Join-Path $package $name
$globalArgs=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',(Join-Path $package 'data/runtime'),'--namespace','asd-alpha-8234c4e','--state-dir',(Join-Path $package 'data/state'))
$result=@{result='started';stages=@();session_content_exported=0;clipboard_calls=0;registry_writes=0;forced_product_ends=0;entrypoint=$name;product_sha256=$expected}
$oldManager=$null;$launched=$false;$manifestChanged=$false;$script:newManager=$null;$script:newWindow=[IntPtr]::Zero;$backup=$null;$shortcutPaths=@();$createdShortcuts=@();$existingHosts=@()
function Stage([string]$name){$script:stage=$name;$result.stages+=@{stage=$name;utc=[DateTime]::UtcNow.ToString('o')}}
function Wait([string]$label,[scriptblock]$check,[int]$seconds=20){$until=[DateTime]::UtcNow.AddSeconds($seconds);do{if(& $check){return};Start-Sleep -Milliseconds 200}while([DateTime]::UtcNow -lt $until);throw ($label+' failed; no automatic replay')}
function Call([string]$exe,[string[]]$arguments){$data=& $exe @globalArgs @arguments 2>$null|Out-String;if($LASTEXITCODE -ne 0){throw 'Local product operation failed'};return $data|ConvertFrom-Json}
function LiveStamp([string]$exe){
 $rows=@(Call $exe @('list','--json'))
 $selected=@($rows|Sort-Object session_id,run_id|ForEach-Object{[ordered]@{session=$_.session_id;run=$_.run_id;backend=$_.backend_id;command=$_.command;title=$_.title;title_revision=$_.title_revision;group=$_.group_id;group_revision=$_.group_revision;position=$_.order;position_revision=$_.order_revision;state=$_.terminal_state}})
 $json=$selected|ConvertTo-Json -Depth 5 -Compress
 return @{value=$json;count=$rows.Count}
}
function ConfigStamp {
 $files=@('ui.toml','launch.toml','last-launch.toml','recent-dirs.toml')
 return (@($files|ForEach-Object{$path=Join-Path $package ('data/state/'+$_);if(Test-Path $path){$_+':'+(Get-FileHash $path).Hash}else{$_+':absent'}})-join '|')
}
function OutputStamp {
 $root=Join-Path $package 'data/state/output-records';if(!(Test-Path $root)){return 'absent'}
 $files=@(Get-ChildItem -LiteralPath $root -File -Recurse)
 if($files.Count -gt 64){throw 'Unexpected output directory size; not audited'}
 return (@($files|Sort-Object FullName|ForEach-Object{$_.FullName+':'+(Get-FileHash $_.FullName).Hash})-join '|')
}
try{
 Stage 'guards'
 if((Get-FileHash $old).Hash.ToLowerInvariant() -cne $oldHash){throw 'Old entry changed'}
 foreach($item in $manifest.files.PSObject.Properties){if((Get-FileHash (Join-Path $package $item.Name)).Hash.ToLowerInvariant() -cne $item.Value){throw 'Existing package payload changed'}}
 if((Get-FileHash (Join-Path $base 'asd-integrated-r3.exe')).Hash.ToLowerInvariant() -cne $expected){throw 'Tested candidate changed'}
 $trial=Get-Content -Raw (Join-Path $base 'daily-migration-trial-summary.json')|ConvertFrom-Json
 if($trial.source_schema -ne 5 -or $trial.clone_schema -ne 9 -or !$trial.history_projection_preserved){throw 'Migration trial is not accepted'}
 $desktop=[DeskFlow]::OpenInputDesktop(0,$false,1);if($desktop -eq [IntPtr]::Zero){throw 'Desktop unavailable'};$null=[DeskFlow]::CloseDesktop($desktop)
 $oldManagers=@(Get-Process -Name ([IO.Path]::GetFileNameWithoutExtension($old)) -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $old})
 if($oldManagers.Count -ne 1){throw 'Old manager not unique; no close'}
 $oldManager=$oldManagers[0];$null=$oldManager.Handle
 $hosts=@([DeskFlow]::HostWindows($hostInfo.terminal.path));if($hosts.Count -ne 1){throw 'Daily terminal window not unique; no close'}
 $window=$hosts[0];[uint32]$hostPID=0;$null=[DeskFlow]::GetWindowThreadProcessId($window,[ref]$hostPID)
 $parent=$oldManager.Id;$ancestor=$false
 for($i=0;$i -lt 32 -and $parent -gt 0;$i++){
  if($parent -eq $hostPID){$ancestor=$true;break}
  $p=Get-CimInstance Win32_Process -Filter ('ProcessId='+$parent) -Property ProcessId,ParentProcessId
  if(!$p){break};$parent=$p.ParentProcessId
 }
 if(!$ancestor){throw 'Daily manager host ancestry not verified; no close'}
 $root=[Windows.Automation.AutomationElement]::FromHandle($window)
 $tabs=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem))
 if($tabs.Count -gt 1){throw 'Daily terminal has other tabs; no close'}
 $associationBefore=[DeskFlow]::AssociationStamp()
 Stage 'wait-for-input-idle'
 Wait 'input idle guard' {return [DeskFlow]::IdleMilliseconds() -ge 45000} 60
 $before=LiveStamp $old
 if(Test-Path $target){throw 'Versioned product target already exists'}
 Copy-Item (Join-Path $base 'asd-integrated-r3.exe') $target
 if((Get-FileHash $target).Hash.ToLowerInvariant() -cne $expected){throw 'Staged product readback mismatch'}
 Stage 'close-only-original-ui'
 $null=[DeskFlow]::SendMessage($window,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)
 if(!$oldManager.WaitForExit(20000)){throw 'Original UI did not close; no forced termination'}
 $result.original_ui_closed_by_own_window=$true
 Stage 'final-private-backup'
 $backup=Join-Path $package 'data/rollback-integrated-20261126'
 if(Test-Path $backup){throw 'Fresh rollback directory required'}
 New-Item -ItemType Directory $backup|Out-Null
 & $target @globalArgs metadata backup (Join-Path $backup 'metadata-v5.sqlite3') 1>$null 2>$null
 if($LASTEXITCODE -ne 0){throw 'Validated final metadata backup failed'}
 Copy-Item $manifestPath (Join-Path $backup 'alpha.json')
 $configBefore=ConfigStamp;$outputsBefore=OutputStamp
 $sourceBefore=(& $target metadata check (Join-Path $package 'data/state/metadata.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0 -or $sourceBefore.schema_version -ne 5){throw 'Unexpected source schema before migration'}
 $null=Call $target @('attention-list')
 $after=LiveStamp $target
 if($before.value -cne $after.value){throw 'Original live identities/labels/order changed; entry not switched'}
 $check=(& $target metadata check (Join-Path $package 'data/state/metadata.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0 -or $check.schema_version -ne 9 -or $check.integrity -cne 'ok'){throw 'Migrated metadata failed validation'}
 Stage 'switch-manifest-and-shortcut'
 $launcherName='start-integrated-74a86959.ps1';$launcherFile=Join-Path $package $launcherName
 if(Test-Path $launcherFile){throw 'Versioned launcher already exists'}
 $launcherBody=@'
param([string]$Root)
$ErrorActionPreference='Stop'
$m=Get-Content -Raw (Join-Path $Root 'alpha.json')|ConvertFrom-Json
foreach($file in $m.files.PSObject.Properties){if((Get-FileHash -LiteralPath (Join-Path $Root $file.Name)).Hash.ToLowerInvariant() -cne $file.Value){throw 'Installed package fingerprint mismatch'}}
& (Join-Path $Root $m.entrypoint) '--pet' '--document-app' 'system-notepad'
exit $LASTEXITCODE
'@
 [IO.File]::WriteAllText($launcherFile,$launcherBody,[Text.UTF8Encoding]::new($true))
 $launcherHash=(Get-FileHash $launcherFile).Hash.ToLowerInvariant()
 $manifest.files|Add-Member -NotePropertyName $launcherName -NotePropertyValue $launcherHash
 $manifest.files|Add-Member -NotePropertyName $name -NotePropertyValue $expected
 $manifest.entrypoint=$name;$manifest.ui_source_commit='845a5b72c37ec6bd3569248812c67084b9e2f614'
 $temp=Join-Path $package 'alpha-integrated.tmp'
 [IO.File]::WriteAllText($temp,($manifest|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($true));Move-Item $temp $manifestPath -Force;$manifestChanged=$true
 $shell=New-Object -ComObject WScript.Shell
 $shortcuts=@(Get-ChildItem ([Environment]::GetFolderPath('Desktop')) -Filter '*Session*Desk*.lnk' -File|Where-Object{$s=$shell.CreateShortcut($_.FullName);$s.TargetPath.StartsWith($package,[StringComparison]::OrdinalIgnoreCase)-or $s.Arguments.IndexOf($package,[StringComparison]::OrdinalIgnoreCase)-ge 0})
 if($shortcuts.Count -gt 5){throw 'Unexpected shortcut count; none changed'}
 $shortcutPaths=@()
 foreach($file in $shortcuts){Copy-Item $file.FullName (Join-Path $backup $file.Name);$shortcutPaths+=$file.FullName}
 if(!$shortcutPaths.Count){$shortcutPaths+=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk.lnk';$createdShortcuts+=$shortcutPaths[0]}
 $launchCode="& ([ScriptBlock]::Create([IO.File]::ReadAllText('"+$launcherFile.Replace("'","''")+"'))) -Root '"+$package.Replace("'","''")+"'"
 $shortcutArguments='-NoProfile -NonInteractive -WindowStyle Hidden -STA -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($launchCode))
 foreach($path in $shortcutPaths){$s=$shell.CreateShortcut($path);$s.TargetPath=$hostInfo.shell.path;$s.Arguments=$shortcutArguments;$s.WorkingDirectory=$package;$s.WindowStyle=7;$s.Description='Agent Session Desk: pet reminders and system Notepad';$s.Save();$readback=$shell.CreateShortcut($path);if($readback.TargetPath -ine $hostInfo.shell.path -or $readback.Arguments -cne $shortcutArguments){throw 'Shortcut readback failed'}}
 Stage 'launch-actual-daily-shortcut'
 $existingHosts=@([DeskFlow]::HostWindows($hostInfo.terminal.path))
 Remove-Item Env:WT_SESSION -ErrorAction SilentlyContinue
 $start=[Diagnostics.ProcessStartInfo]::new();$start.FileName=$shortcutPaths[0];$start.UseShellExecute=$true;$null=[Diagnostics.Process]::Start($start);$launched=$true
 Wait 'one hosted manager and live pet' {
  $ps=@(Get-Process -Name ([IO.Path]::GetFileNameWithoutExtension($target)) -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target})
  if($ps.Count -ne 1){return $false}
  $pets=@([DeskFlow]::Find('ASD Blocking Reminders',$ps[0].Id));if($pets.Count -ne 1){return $false}
  $script:newManager=$ps[0];$null=$script:newManager.Handle;return $true
 }
 Wait 'new daily terminal window' {$new=@([DeskFlow]::HostWindows($hostInfo.terminal.path)|Where-Object{$_ -notin $existingHosts});if($new.Count -eq 1){$script:newWindow=$new[0];return $true};return $false}
 Start-Sleep -Seconds 3
 if($script:newManager.HasExited){throw 'Daily manager exited after startup'}
 if($after.count -gt 0){Wait 'one active native terminal client' {$connections=@(Get-NetTCPConnection -OwningProcess $script:newManager.Id -State Established -ErrorAction SilentlyContinue|Where-Object{$_.LocalAddress -in @('127.0.0.1','::1')});return $connections.Count -eq 1}}
 $final=LiveStamp $target
 if($before.value -cne $final.value){throw 'Live identities/labels/order changed after reopen'}
 if($configBefore -cne (ConfigStamp)-or $outputsBefore -cne (OutputStamp)){throw 'Launch/UI preferences or output records changed'}
 if($associationBefore -cne [DeskFlow]::AssociationStamp()){throw 'Effective Markdown association changed'}
 $result.result='updated-and-running';$result.schema_before=5;$result.schema_after=9;$result.private_backup_validated=$true;$result.live_sessions=$final.count;$result.original_live_projection_preserved=$true;$result.preferences_and_output_records_preserved=$true;$result.effective_md_association_preserved=$true;$result.shortcut_count=$shortcutPaths.Count;$result.manager_count=1;$result.live_pet=$true;$result.default_windows_terminal_handoff_verified=$true;$result.owner_window_binding_present=([DeskFlow]::OwnerBindings($script:newWindow)-gt 0);if(!$result.owner_window_binding_present){throw 'Daily foreground owner binding unavailable'}
 }catch{
 $result.result='failed';$result.stage=$script:stage;$result.error=$_.Exception.Message;$result.manifest_changed=$manifestChanged;$result.new_launch_attempted=$launched;$result.automatic_replay=$false
 # Roll back only after all new UI processes end and the known private backup
 # exists. Preserve the upgraded file itself; never delete or overwrite it.
 if($oldManager -and $oldManager.HasExited -and $backup -and (Test-Path (Join-Path $backup 'metadata-v5.sqlite3'))){
  try{
   if($script:newWindow -ne [IntPtr]::Zero){$root=[Windows.Automation.AutomationElement]::FromHandle($script:newWindow);$tabs=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem));if($tabs.Count -gt 1){throw 'New terminal has additional tabs; rollback paused'};$null=[DeskFlow]::SendMessage($script:newWindow,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)}
   if($script:newManager -and !$script:newManager.HasExited){$null=$script:newManager.WaitForExit(15000)}
   if(@(Get-Process -Name ([IO.Path]::GetFileNameWithoutExtension($target)) -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target}).Count){throw 'New UI still running; metadata rollback paused'}
   $now=LiveStamp $target
   if($now.value -cne $before.value -or $configBefore -cne (ConfigStamp)-or $outputsBefore -cne (OutputStamp)){throw 'State changed during update; preserve current data and pause rollback'}
   $database=Join-Path $package 'data/state/metadata.sqlite3'
   $retained=Join-Path $backup 'metadata-after-update-retained.sqlite3'
   if(Test-Path $retained){throw 'Retained database destination already exists'}
   Move-Item $database $retained
   & $target @globalArgs metadata restore (Join-Path $backup 'metadata-v5.sqlite3') 1>$null 2>$null
   if($LASTEXITCODE -ne 0){if(!(Test-Path $database)){Move-Item $retained $database};throw 'Metadata restore failed; updated database retained'}
   Copy-Item (Join-Path $backup 'alpha.json') $manifestPath -Force
   foreach($path in $shortcutPaths){$saved=Join-Path $backup ([IO.Path]::GetFileName($path));if(Test-Path $saved){Copy-Item $saved $path -Force}elseif($path -in $createdShortcuts -and (Test-Path $path)){Remove-Item -LiteralPath $path}}
   Remove-Item Env:WT_SESSION -ErrorAction SilentlyContinue
   $start=[Diagnostics.ProcessStartInfo]::new();$start.FileName=$old;$start.WorkingDirectory=$package;$start.UseShellExecute=$true;$null=[Diagnostics.Process]::Start($start)
   $result.rollback_restored_v5_and_old_entry=$true;$result.updated_database_retained_locally=$true
  }catch{$result.rollback_paused=$true;$result.rollback_reason=$_.Exception.Message}
 }
}
$result|ConvertTo-Json -Depth 6|Set-Content $receipt -Encoding utf8
