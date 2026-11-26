$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$package=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e'
$out=Join-Path $base 'daily-integrated-repair-20261126'
if(Test-Path $out){throw 'Fresh repair directory required'}
New-Item -ItemType Directory $out|Out-Null
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $base 'desktop-api.ps1'))))
$target=Join-Path $package 'asd-integrated-74a86959.exe'
$source=Join-Path $base 'daily-retained-private-compare-20261126/metadata.sqlite3'
$state=Join-Path $package 'data/state';$database=Join-Path $state 'metadata.sqlite3'
$hostInfo=Get-Content -Raw (Join-Path $package 'local-host.json')|ConvertFrom-Json
$g=@('--terminal-bundle',(Join-Path $package 'terminal-runtime'),'--runtime-dir',(Join-Path $package 'data/runtime'),'--namespace','asd-alpha-8234c4e')
$r=@{result='started';stages=@();session_content_exported=0;clipboard_calls=0;registry_writes=0;forced_product_ends=0;automatic_launch_replay=$false;automatic_metadata_rollback=$false}
function Stage([string]$name){$script:stage=$name;$r.stages+=@{stage=$name;utc=[DateTime]::UtcNow.ToString('o')}}
function Wait([string]$label,[scriptblock]$condition,[int]$seconds=30){$until=[DateTime]::UtcNow.AddSeconds($seconds);do{if(& $condition){return};Start-Sleep -Milliseconds 250}while([DateTime]::UtcNow -lt $until);throw ($label+' failed; observe only, no automatic replay or rollback')}
function ProductJSON([string]$dir,[string[]]$arguments){$text=(& $target @g '--state-dir' $dir @arguments 2>$null|Out-String);if($LASTEXITCODE -ne 0){throw 'Local product check failed'};return $text}
function HistoryProjection([string]$dir){$rows=(ProductJSON $dir @('history','--limit','100'))|ConvertFrom-Json;$selected=@($rows|ForEach-Object{$s=$_.session;[ordered]@{session=$s.session_id;run=$s.run_id;backend=$s.backend_id;command=$s.command;title=$s.title;title_revision=$s.title_revision;group=$s.group_id;group_revision=$s.group_revision;position=$s.order;position_revision=$s.order_revision}}|Sort-Object {$_.session},{$_.run});return ConvertTo-Json -InputObject $selected -Depth 5 -Compress}
function LiveIdentity([string]$dir){$rows=(ProductJSON $dir @('list','--json'))|ConvertFrom-Json;$selected=@($rows|ForEach-Object{[ordered]@{session=$_.session_id;run=$_.run_id;backend=$_.backend_id;command=$_.command}}|Sort-Object {$_.session},{$_.run});return ConvertTo-Json -InputObject $selected -Depth 5 -Compress}
function ConfigStamp {$items=@('ui.toml','launch.toml','last-launch.toml','recent-dirs.toml');return (@($items|ForEach-Object{$path=Join-Path $state $_;if(Test-Path $path){$_+':'+(Get-FileHash $path).Hash}else{$_+':absent'}})-join '|')}
function OutputStamp {$root=Join-Path $state 'output-records';if(!(Test-Path $root)){return 'absent'};$files=@(Get-ChildItem -LiteralPath $root -File -Recurse);if($files.Count -gt 64){throw 'Output count exceeds review bound'};return (@($files|Sort-Object FullName|ForEach-Object{$_.FullName+':'+(Get-FileHash $_.FullName).Hash})-join '|')}
try{
 Stage 'guards-and-private-source'
 if((Get-FileHash $target).Hash.ToLowerInvariant() -cne '74a86959355ba37ea641d08c8664602fd8daa8e8313c1b621fffe4986fe31d99'){throw 'Tested product changed'}
 $proof=Get-Content -Raw (Join-Path $base 'daily-retained-comparison-summary.json')|ConvertFrom-Json
 if(!$proof.retained_original_history_projection_preserved -or !$proof.private_snapshot_validated){throw 'Retained source proof missing'}
 $check=(& $target metadata check $source 2>$null|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0 -or $check.schema_version -ne 9 -or $check.integrity -cne 'ok'){throw 'Retained source invalid'}
 $sourceState=Split-Path $source
 $originalProjection=HistoryProjection $sourceState
 $originalRows=(ProductJSON $sourceState @('history','--limit','100'))|ConvertFrom-Json
 if($originalRows.Count -ne 6){throw 'Unexpected original source count'}
 $originalLive=LiveIdentity $state
 $configs=ConfigStamp;$outputs=OutputStamp;$association=[DeskFlow]::AssociationStamp()
 # Only private hashes and local projections are retained for independent audit.
 @{history=$originalProjection;live=$originalLive;configs=$configs;outputs=$outputs;association=$association}|ConvertTo-Json -Depth 3|Set-Content (Join-Path $out 'baseline-private.json') -Encoding utf8
 $ps=@(Get-Process -Name 'asd-integrated-74a86959' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target});if($ps.Count -ne 1){throw 'Current manager not unique'}
 $manager=$ps[0];$null=$manager.Handle
 $hosts=@([DeskFlow]::HostWindows($hostInfo.terminal.path));if($hosts.Count -ne 1 -or [DeskFlow]::OwnerBindings($hosts[0]) -ne 1){throw 'Bound terminal window not unique'}
 $window=$hosts[0];[uint32]$hostPID=0;$null=[DeskFlow]::GetWindowThreadProcessId($window,[ref]$hostPID)
 $parent=$manager.Id;$ancestor=$false
 for($i=0;$i -lt 32 -and $parent -gt 0;$i++){if($parent -eq $hostPID){$ancestor=$true;break};$p=Get-CimInstance Win32_Process -Filter ('ProcessId='+$parent) -Property ProcessId,ParentProcessId;if(!$p){break};$parent=$p.ParentProcessId}
 if(!$ancestor){throw 'Current host ancestry not verified'}
 Stage 'wait-for-input-idle'
 Wait 'input idle guard' {return [DeskFlow]::IdleMilliseconds() -ge 45000} 60
 $root=[Windows.Automation.AutomationElement]::FromHandle($window)
 $tabs=$root.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem))
 if($tabs.Count -gt 1){throw 'Other terminal tabs exist; no close'}
 Stage 'close-only-current-ui-and-settle'
 $null=[DeskFlow]::SendMessage($window,0x10,[UIntPtr]::Zero,[IntPtr]::Zero)
 if(!$manager.WaitForExit(30000)){throw 'Current UI still running; no forced end'}
 for($i=0;$i -lt 32;$i++){if(@(Get-Process -Name 'asd-integrated-74a86959' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target}).Count){throw 'A product launch is still active; do not replace metadata'};Start-Sleep -Milliseconds 250}
 foreach($suffix in @('-wal','-shm')){if(Test-Path ($database+$suffix)){throw 'SQLite journal remains; do not move database'}}
 $r.original_ui_ended=$true
 Stage 'preserve-current-and-restore-validated-source'
 $null=ProductJSON $state @('metadata','backup',(Join-Path $out 'current-before-repair.sqlite3'))
 if((ConfigStamp)-cne $configs -or (OutputStamp)-cne $outputs -or (LiveIdentity $state)-cne $originalLive){throw 'State changed before restore'}
 foreach($suffix in @('-wal','-shm')){if(Test-Path ($database+$suffix)){throw 'SQLite journal remains after comparison'}}
 Move-Item $database (Join-Path $out 'current-database-retained.sqlite3')
 $null=ProductJSON $state @('metadata','restore',$source)
 if((HistoryProjection $state)-cne $originalProjection){throw 'Restored history differs; do not launch'}
 $r.original_history_restored=$true
 Stage 'launch-actual-shortcut-once'
 $shell=New-Object -ComObject WScript.Shell
 $links=@(Get-ChildItem ([Environment]::GetFolderPath('Desktop')) -Filter '*Session*Desk*.lnk' -File|Where-Object{$s=$shell.CreateShortcut($_.FullName);try{$body=[Text.Encoding]::Unicode.GetString([Convert]::FromBase64String(($s.Arguments -split ' ')[-1]));return $body.IndexOf((Join-Path $package 'start-integrated-74a86959.ps1'),[StringComparison]::OrdinalIgnoreCase)-ge 0}catch{return $false}})
 if($links.Count -ne 1){throw 'Updated daily shortcut not unique'}
 $existing=@([DeskFlow]::HostWindows($hostInfo.terminal.path))
 Remove-Item Env:WT_SESSION -ErrorAction SilentlyContinue
 $start=[Diagnostics.ProcessStartInfo]::new();$start.FileName=$links[0].FullName;$start.UseShellExecute=$true;$null=[Diagnostics.Process]::Start($start)
 $r.shortcut_launches=1
 $script:newManager=$null;$script:newWindow=[IntPtr]::Zero
 Wait 'one manager, pet and bound terminal' {
  $items=@(Get-Process -Name 'asd-integrated-74a86959' -ErrorAction SilentlyContinue|Where-Object{$_.Path -ieq $target})
  if($items.Count -ne 1){return $false}
  if(@([DeskFlow]::Find('ASD Blocking Reminders',$items[0].Id)).Count -ne 1){return $false}
  $wins=@([DeskFlow]::HostWindows($hostInfo.terminal.path)|Where-Object{$_ -notin $existing -and [DeskFlow]::OwnerBindings($_) -eq 1})
  if($wins.Count -ne 1){return $false}
  $script:newManager=$items[0];$null=$script:newManager.Handle;$script:newWindow=$wins[0];return $true
 } 90
 Stage 'verify-original-metadata-and-running-workspace'
 Start-Sleep -Seconds 5
 if($script:newManager.HasExited){throw 'New UI exited'}
 if((HistoryProjection $state)-cne $originalProjection){throw 'Original history projection changed after startup'}
 if((LiveIdentity $state)-cne $originalLive){throw 'Original live identities changed'}
 if((ConfigStamp)-cne $configs -or (OutputStamp)-cne $outputs){throw 'Preferences or outputs changed'}
 if([DeskFlow]::AssociationStamp()-cne $association){throw 'Effective association changed'}
 $check=(& $target metadata check $database 2>$null|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0 -or $check.schema_version -ne 9 -or $check.integrity -cne 'ok'){throw 'Final structural metadata invalid'}
 $connections=@(Get-NetTCPConnection -OwningProcess $script:newManager.Id -State Established -ErrorAction SilentlyContinue|Where-Object{$_.LocalAddress -in @('127.0.0.1','::1')})
 $r.result='repaired-updated-and-running';$r.history_projection_preserved=$true;$r.original_history_count=6;$r.live_identity_projection_preserved=$true;$r.configs_and_outputs_preserved=$true;$r.effective_association_preserved=$true;$r.manager_count=1;$r.pet_count=1;$r.bound_terminal_count=1;$r.native_client_connections=$connections.Count;$r.current_schema=9;$r.current_integrity='ok';$r.current_data_retained_locally=$true
}catch{$r.result='repair-stopped';$r.stage=$script:stage;$r.error=$_.Exception.Message}
$r|ConvertTo-Json -Depth 6|Set-Content (Join-Path $out 'result.json') -Encoding utf8
