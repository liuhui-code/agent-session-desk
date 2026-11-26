$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$out=Join-Path $base 'pet-fix-daily-20261126';if(Test-Path $out){throw 'Fresh handoff directory required'};New-Item -ItemType Directory $out|Out-Null
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$previousDpi=[DeskFlow]::SetThreadDpiAwarenessContext([IntPtr](-4));if($previousDpi -eq [IntPtr]::Zero){throw 'DPI context unavailable'}
$expected='1897bc9dc73fe519b256a073b368695d3e6accc3495933189ba9d68d9d167105';$candidate=Join-Path $base 'asd-pet-fix.exe'
$name='asd-pet-1897bc9d.exe';$target=Join-Path $package $name
$manifestPath=Join-Path $package 'alpha.json';$manifest=Get-Content -Raw $manifestPath|ConvertFrom-Json
$old=Join-Path $package $manifest.entrypoint
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$state=Join-Path $package 'data/state'
$g=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',(Join-Path $package 'data/runtime'),'--namespace','asd-alpha-8234c4e','--state-dir',$state)
$r=@{result='started';stages=@();registry_writes=0;clipboard_calls=0;session_content_exported=0;terminal_key_actions=0;forced_product_ends=0;automatic_launch_replay=$false;metadata_restore_attempts=0;metadata_migration_attempts=0}
$script:note=$null;$script:noteWindow=[IntPtr]::Zero;$script:newManager=$null;$script:newPet=[IntPtr]::Zero;$script:newWindow=[IntPtr]::Zero
$cursor=[DeskFlow+P]::new();$null=[DeskFlow]::GetCursorPos([ref]$cursor)
function Stage([string]$name){$script:stage=$name;$r.stages+=@{stage=$name;utc=[DateTime]::UtcNow.ToString('o')}}
function Wait([string]$label,[scriptblock]$condition,[int]$seconds=30){$until=[DateTime]::UtcNow.AddSeconds($seconds);do{if(& $condition){return};Start-Sleep -Milliseconds 200}while([DateTime]::UtcNow -lt $until);throw ($label+' failed; observe only, no automatic replay or metadata rollback')}
function JSONText([string]$exe,[string[]]$arguments){$text=(& $exe @g @arguments 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'Local product operation failed'};return $text}
function Live([string]$exe){$rows=(JSONText $exe @('list','--json'))|ConvertFrom-Json;$values=@($rows|ForEach-Object{[ordered]@{session=$_.session_id;run=$_.run_id;backend=$_.backend_id;command=$_.command;title=$_.title;title_revision=$_.title_revision;group=$_.group_id;group_revision=$_.group_revision;order=$_.order;order_revision=$_.order_revision}}|Sort-Object {$_.session},{$_.run});return ConvertTo-Json -InputObject $values -Depth 5 -Compress}
function History([string]$exe){$rows=(JSONText $exe @('history','--limit','100'))|ConvertFrom-Json;$values=@($rows|ForEach-Object{$s=$_.session;[ordered]@{session=$s.session_id;run=$s.run_id;backend=$s.backend_id;command=$s.command;title=$s.title;title_revision=$s.title_revision;group=$s.group_id;group_revision=$s.group_revision}}|Sort-Object {$_.session},{$_.run});return ConvertTo-Json -InputObject $values -Depth 5 -Compress}
function Configs {$files=@('ui.toml','launch.toml','last-launch.toml','recent-dirs.toml');return (@($files|ForEach-Object{$p=Join-Path $state $_;if(Test-Path $p){$_+':'+(Get-FileHash $p).Hash}else{$_+':absent'}})-join '|')}
function Outputs {$root=Join-Path $state 'output-records';if(!(Test-Path $root)){return 'absent'};$files=@(Get-ChildItem -LiteralPath $root -File -Recurse);if($files.Count -gt 64){throw 'Output bound exceeded'};return (@($files|Sort-Object FullName|ForEach-Object{$_.FullName+':'+(Get-FileHash $_.FullName).Hash})-join '|')}
try{
 Stage 'guards'
 if($manifest.entrypoint -cne 'asd-integrated-74a86959.exe' -or (Get-FileHash $old).Hash.ToLowerInvariant() -cne '74a86959355ba37ea641d08c8664602fd8daa8e8313c1b621fffe4986fe31d99'){throw 'Daily baseline changed'}
 foreach($item in $manifest.files.PSObject.Properties){if((Get-FileHash (Join-Path $package $item.Name)).Hash.ToLowerInvariant() -cne $item.Value){throw 'Current package hash mismatch'}}
 if((Get-FileHash $candidate).Hash.ToLowerInvariant() -cne $expected){throw 'Tested product changed'}
 if(Test-Path $target){throw 'Versioned product already exists'}
 $native=Get-Content -Raw (Join-Path $base 'pet-fix-native-20261126/tests.json')|ConvertFrom-Json
 $desktop=Get-Content -Raw (Join-Path $base 'pet-fix-desktop-20261126/result.json')|ConvertFrom-Json
 if($native.Count -ne 2 -or @($native|Where-Object{$_.exit_code -ne 0 -or $_.top_fail -ne 0 -or $_.top_skip -ne 0}).Count -or $desktop.result -cne 'verified' -or !$desktop.idle_click_restored_minimized_terminal){throw 'Required native/desktop evidence not passed'}
 $ps=@(Get-Process -Name 'asd-integrated-74a86959' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $old});if($ps.Count -ne 1){throw 'Original daily UI not unique'}
 $manager=$ps[0];$null=$manager.Handle
 $hosts=@([DeskFlow]::HostWindows($hostInfo.terminal.path));if($hosts.Count -ne 1 -or [DeskFlow]::OwnerBindings($hosts[0]) -ne 1){throw 'Original bound host not unique'}
 $window=$hosts[0];[uint32]$hostPID=0;$null=[DeskFlow]::GetWindowThreadProcessId($window,[ref]$hostPID)
 $ancestor=$false;$parent=$manager.Id
 for($i=0;$i -lt 32 -and $parent -gt 0;$i++){if($parent -eq $hostPID){$ancestor=$true;break};$p=Get-CimInstance Win32_Process -Filter ('ProcessId='+$parent) -Property ProcessId,ParentProcessId;if(!$p){break};$parent=$p.ParentProcessId}
 if(!$ancestor){throw 'Original host ancestry unavailable'}
 $check=(& $old metadata check (Join-Path $state 'metadata.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0 -or $check.schema_version -ne 9 -or $check.integrity -cne 'ok'){throw 'Existing metadata invalid; no migration'}
 Stage 'wait-for-input-idle'
 Wait 'input idle guard' {[DeskFlow]::IdleMilliseconds()-ge 45000} 60
 $root=[Windows.Automation.AutomationElement]::FromHandle($window)
 $tabs=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem))
 if($tabs.Count -gt 1){throw 'Other native tabs exist; no close'}
 $liveBefore=Live $old;$historyBefore=History $old;$configBefore=Configs;$outputBefore=Outputs;$associationBefore=[DeskFlow]::AssociationStamp()
 @{live=$liveBefore;history=$historyBefore;configs=$configBefore;outputs=$outputBefore;association=$associationBefore}|ConvertTo-Json -Depth 3|Set-Content (Join-Path $out 'baseline-private.json') -Encoding utf8
 Copy-Item $manifestPath (Join-Path $out 'alpha-before-private.json')
 Copy-Item $candidate $target
 if((Get-FileHash $target).Hash.ToLowerInvariant() -cne $expected){throw 'Staged candidate mismatch'}
 Stage 'close-only-original-ui'
 $null=[DeskFlow]::SendMessage($window,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)
 if(!$manager.WaitForExit(30000)){throw 'Original UI did not exit; no forced end'}
 $r.old_ui_ended=$true
 Stage 'snapshot-only-no-database-replacement'
 $null=JSONText $old @('metadata','backup',(Join-Path $out 'metadata-v9-before.sqlite3'))
 if((Live $old)-cne $liveBefore -or (History $old)-cne $historyBefore -or (Configs)-cne $configBefore -or (Outputs)-cne $outputBefore){throw 'Original state changed before entry switch'}
 $manifest.files|Add-Member -NotePropertyName $name -NotePropertyValue $expected
 $manifest.entrypoint=$name;$manifest.ui_source_commit='fbd9bfbb351296c4856644f6af022507322a1ba2'
 $temporary=Join-Path $package 'alpha-pet-fix.tmp';if(Test-Path $temporary){throw 'Manifest staging file exists'}
 [IO.File]::WriteAllText($temporary,($manifest|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($true));Move-Item $temporary $manifestPath -Force
 $r.manifest_switched=$true
 Stage 'launch-existing-daily-shortcut-once'
 $shell=New-Object -ComObject WScript.Shell
 $links=@(Get-ChildItem ([Environment]::GetFolderPath('Desktop')) -Filter '*Session*Desk*.lnk' -File|Where-Object{$s=$shell.CreateShortcut($_.FullName);try{$body=[Text.Encoding]::Unicode.GetString([Convert]::FromBase64String(($s.Arguments -split ' ')[-1]));return $body.IndexOf((Join-Path $package 'start-integrated-74a86959.ps1'),[StringComparison]::OrdinalIgnoreCase)-ge 0}catch{return $false}})
 if($links.Count -ne 1){throw 'Existing daily shortcut not unique'}
 $existing=@([DeskFlow]::HostWindows($hostInfo.terminal.path))
 Remove-Item Env:WT_SESSION -ErrorAction SilentlyContinue
 $info=[Diagnostics.ProcessStartInfo]::new();$info.FileName=$links[0].FullName;$info.UseShellExecute=$true;$null=[Diagnostics.Process]::Start($info);$r.shortcut_launches=1
 Wait 'new manager, pet and original host binding' {
  $items=@(Get-Process -Name 'asd-pet-1897bc9d' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target});if($items.Count -ne 1){return $false}
  $pets=@([DeskFlow]::Find('ASD Blocking Reminders',$items[0].Id));if($pets.Count -ne 1){return $false}
  $wins=@([DeskFlow]::HostWindows($hostInfo.terminal.path)|Where-Object{$_ -notin $existing -and [DeskFlow]::OwnerBindings($_)-eq 1});if($wins.Count -ne 1){return $false}
  $script:newManager=$items[0];$null=$script:newManager.Handle;$script:newPet=$pets[0];$script:newWindow=$wins[0];return $true
 } 90
 Stage 'actual-idle-click-from-owned-notepad'
 if([DeskFlow]::Title($script:newPet)-cne 'ASD Blocking Reminders · idle · 0'){throw 'Daily reminder is not idle; preserve pending without injected test input'}
 $doc=Join-Path $out 'focus-check.md';[IO.File]::WriteAllText($doc,'# ASD pet focus check',[Text.UTF8Encoding]::new($false))
 $info=[Diagnostics.ProcessStartInfo]::new();$info.FileName=Join-Path ([Environment]::SystemDirectory) 'notepad.exe';$info.UseShellExecute=$false;$info.ArgumentList.Add($doc)
 $script:note=[Diagnostics.Process]::Start($info);$null=$script:note.Handle
 Wait 'controlled note window' {$matches=@([DeskFlow]::Find('',$script:note.Id));if($matches.Count -eq 1){$script:noteWindow=$matches[0];return $true};return $false}
 $null=[DeskFlow]::SetForegroundWindow($script:noteWindow)
 Wait 'controlled note foreground' {[DeskFlow]::GetForegroundWindow()-eq $script:noteWindow}
 foreach($key in @(1,2,4,16,17,18,91,92)){if(([DeskFlow]::GetAsyncKeyState($key)-band 32768)-ne 0){throw 'User key/button held; no mouse replay'}}
 $rect=[DeskFlow+R]::new();if(![DeskFlow]::GetWindowRect($script:newPet,[ref]$rect)){throw 'Owned pet rectangle unavailable'}
 $script:lastX=[int](($rect.Left+$rect.Right)/2);$script:lastY=[int]($rect.Top+($rect.Bottom-$rect.Top)*0.92)
 if([DeskFlow]::Move($script:lastX,$script:lastY)-ne 1 -or [DeskFlow]::Click()-ne 2){throw 'Actual pet click incomplete; no replay'}
 Wait 'daily pet restores original terminal foreground' {[DeskFlow]::GetForegroundWindow()-eq $script:newWindow} 8
 $r.actual_idle_click_restored_original_terminal=$true
 Stage 'verify-preserved-state'
 if((Live $target)-cne $liveBefore -or (History $target)-cne $historyBefore){throw 'Original identity/label/group/order projection changed'}
 if((Configs)-cne $configBefore -or (Outputs)-cne $outputBefore){throw 'Configuration or output hashes changed'}
 if([DeskFlow]::AssociationStamp()-cne $associationBefore){throw 'Effective Markdown association changed'}
 $check=(& $target metadata check (Join-Path $state 'metadata.sqlite3') 2>$null|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0 -or $check.schema_version -ne 9 -or $check.integrity -cne 'ok'){throw 'Final metadata invalid'}
 $rows=(JSONText $target @('list','--json'))|ConvertFrom-Json
 $r.result='updated-and-running';$r.original_live_count=$rows.Count;$r.history_and_live_projection_preserved=$true;$r.configuration_and_output_hashes_preserved=$true;$r.effective_markdown_association_preserved=$true;$r.schema_before=9;$r.schema_after=9;$r.integrity='ok';$r.product_sha256=$expected;$r.entrypoint=$name;$r.manager_count=1;$r.pet_count=1;$r.bound_host_count=1
}catch{$r.result='stopped';$r.stage=$script:stage;$r.error=$_.Exception.Message}
finally{
 if($script:note -and !$script:note.HasExited -and $script:noteWindow -ne [IntPtr]::Zero){$null=[DeskFlow]::SendMessage($script:noteWindow,0x10,[UIntPtr]::Zero,[IntPtr]::Zero);$r.controlled_note_closed=$script:note.WaitForExit(10000)}
 $now=[DeskFlow+P]::new();$null=[DeskFlow]::GetCursorPos([ref]$now)
 if($null -ne $script:lastX -and $now.X -eq $script:lastX -and $now.Y -eq $script:lastY){$r.cursor_restored=([DeskFlow]::Move($cursor.X,$cursor.Y)-eq 1)}
 $null=[DeskFlow]::SetThreadDpiAwarenessContext($previousDpi)
 $r|ConvertTo-Json -Depth 6|Set-Content (Join-Path $out 'result.json') -Encoding utf8
}
