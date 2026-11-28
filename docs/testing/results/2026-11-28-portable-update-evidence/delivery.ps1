$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$p=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e';$o=Join-Path $b 'update-delivery-db63c5a'
if(Test-Path $o){throw 'Fresh delivery required'};New-Item -ItemType Directory $o|Out-Null
$r=@{result='started';forced=0;terminal_input=0;clipboard_apis=0;stages=@()}
function Stage([string]$name){$r.stages+=@{name=$name;at=[DateTime]::UtcNow.ToString('o')}}
try{
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $b 'desktop-api.ps1'))))
$m=Get-Content -Raw (Join-Path $p 'alpha.json')|ConvertFrom-Json;$h=Get-Content -Raw (Join-Path $p 'local-host.json')|ConvertFrom-Json
$old=Join-Path $p $m.entrypoint;$new=Join-Path $p 'asd-updater-db63c5a.exe';$source=Join-Path (Join-Path $env:USERPROFILE 'asd-update-20261128') 'asd-update-product.exe';$expected='069e4d2ed87ed08696ea12ee0cfddb0ab0a0d42efd44534d5324a84a207970e9'
if($m.entrypoint-cne 'asd-desk-44aa154.exe'-or (Get-FileHash $old).Hash.ToLowerInvariant()-cne '7e45eaf62b813d4b0c6e3943e29a146430c0202d6f68690df243fa00a01f27e7'){throw 'Daily baseline changed'}
if((Test-Path $new)-or (Get-FileHash $source).Hash.ToLowerInvariant()-cne $expected){throw 'Candidate guard'}
if((Get-FileHash (Join-Path $p 'terminal-runtime/psmux.exe')).Hash.ToLowerInvariant()-cne 'ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b'){throw 'Backend guard'}
$ps=@(Get-Process|Where-Object{$_.Path-ceq $old});if($ps.Count-ne 1){throw 'Original manager count'};$op=$ps[0];$null=$op.Handle
$g=@('--terminal-bundle',(Join-Path $p 'terminal-runtime'),'--runtime-dir',(Join-Path $p 'data/runtime'),'--namespace','asd-alpha-8234c4e','--state-dir',(Join-Path $p 'data/state'))
function Live($e){$t=& $e @g list --json 2>$null;if($LASTEXITCODE-ne 0){throw 'Live observation failed'};return ConvertTo-Json -InputObject @(($t|Out-String|ConvertFrom-Json)|ForEach-Object{[ordered]@{s=$_.session_id;r=$_.run_id;b=$_.backend_id}}|Sort-Object s,r) -Compress}
function Configs{ $q=Join-Path $p 'data/state';return (@('ui.toml','launch.toml','last-launch.toml','recent-dirs.toml','keybindings.toml')|ForEach-Object{$f=Join-Path $q $_;if(Test-Path $f){$_+':'+(Get-FileHash $f).Hash}else{$_+':absent'}})-join '|'}
if($op.Id-ne 21436-or $op.StartTime.ToUniversalTime().ToString('o')-cne '2026-11-28T01:52:55.0438966Z'){throw 'Original manager identity changed'}
foreach($hostImage in @($h.terminal,$h.shell)){if((Get-FileHash $hostImage.path).Hash.ToLowerInvariant()-cne $hostImage.sha256.ToLowerInvariant()){throw 'Host image guard'}}
Stage 'guards';$before=Live $old;$configs=Configs;[IO.File]::WriteAllText((Join-Path $o 'live-before-private.json'),$before);[IO.File]::WriteAllText((Join-Path $o 'configs-before-private.txt'),$configs);$r.before_saved=$true;$owned=@()
foreach($w in @([DeskFlow]::HostWindows($h.terminal.path))){[uint32]$hp=0;$null=[DeskFlow]::GetWindowThreadProcessId($w,[ref]$hp);$parent=$op.Id;for($i=0;$i-lt 16-and $parent-gt 0;$i++){if($parent-eq $hp){$owned+=$w;break};$q=Get-CimInstance Win32_Process -Filter ('ProcessId='+$parent) -Property ParentProcessId;if(!$q){break};$parent=$q.ParentProcessId}}
if($owned.Count-ne 1){throw 'Owned host not unique'};$w=$owned[0];if([DeskFlow]::OwnerBindings($w)-ne 1){throw 'Bound owner not unique'}
$el=[Windows.Automation.AutomationElement]::FromHandle($w);$tabs=$el.FindAll([Windows.Automation.TreeScope]::Descendants,[Windows.Automation.PropertyCondition]::new([Windows.Automation.AutomationElement]::ControlTypeProperty,[Windows.Automation.ControlType]::TabItem));if($tabs.Count-ne 1){throw 'Other tabs'}
$sh=New-Object -ComObject WScript.Shell;$links=@(Get-ChildItem ([Environment]::GetFolderPath('Desktop')) -Filter '*Session*Desk*.lnk'|Where-Object{$s=$sh.CreateShortcut($_);if($s.Arguments-match '-EncodedCommand\s+(\S+)'){$v=[Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($Matches[1]));$v.Contains($p)}else{$false}});if($links.Count-ne 1){throw 'Shortcut not unique'}
Copy-Item (Join-Path $p 'alpha.json') (Join-Path $o 'alpha-before-private.json');Copy-Item $source $new
Stage 'close-ui';$null=[DeskFlow]::SendMessage($w,0x10,[UIntPtr]::Zero,[IntPtr]::Zero);if(!$op.WaitForExit(20000)){throw 'Original UI did not close; no force'}
Stage 'metadata-backup';$backup=Join-Path $o 'metadata-v11-private.sqlite3';$null=& $old @g metadata backup $backup 2>$null;if($LASTEXITCODE-ne 0){throw 'Metadata backup failed'}
$check=(& $old metadata check $backup 2>$null|Out-String|ConvertFrom-Json);if($LASTEXITCODE-ne 0-or $check.schema_version-ne 11-or $check.integrity-cne 'ok'){throw 'Backup validation failed'};$r.backup_schema=11;$r.backup_valid=$true
$m.entrypoint='asd-updater-db63c5a.exe';$m.ui_source_commit='db63c5a';$m.files|Add-Member -NotePropertyName $m.entrypoint -NotePropertyValue $expected
$tmp=Join-Path $p 'alpha-updater-db63c5a.tmp';if(Test-Path $tmp){throw 'Staging manifest exists'};[IO.File]::WriteAllText($tmp,($m|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($true));Move-Item $tmp (Join-Path $p 'alpha.json') -Force;$r.manifest_switched=$true
Stage 'launch-once';$si=[Diagnostics.ProcessStartInfo]::new();$si.FileName=$links[0];$si.UseShellExecute=$true;$null=[Diagnostics.Process]::Start($si)
$until=[DateTime]::UtcNow.AddSeconds(180);$np=$null;do{$matches=@(Get-Process|Where-Object{$_.Path-ceq $new});if($matches.Count-eq 1-and @([DeskFlow]::Find('ASD Blocking Reminders',$matches[0].Id)).Count-eq 1){$np=$matches[0];break};Start-Sleep -Milliseconds 200}while([DateTime]::UtcNow-lt $until)
if(!$np){throw 'New manager timeout; no replay'}
Stage 'verify';$after=Live $new;[IO.File]::WriteAllText((Join-Path $o 'live-after-private.json'),$after);[IO.File]::WriteAllText((Join-Path $o 'configs-after-private.txt'),(Configs));if($before-cne $after){throw 'Live identity set changed'};if($configs-cne (Configs)){throw 'Configuration hashes changed'}
$check=(& $new metadata check (Join-Path $p 'data/state/metadata.sqlite3') 2>$null|Out-String|ConvertFrom-Json);if($LASTEXITCODE-ne 0-or $check.schema_version-ne 11-or $check.integrity-cne 'ok'){throw 'Final metadata check failed'}
$r.result='updated-running';$r.live_preserved=$true;$r.config_preserved=$true;$r.final_schema=11;$r.hash=$expected;$r.source='db63c5a';$r.manager_pid=$np.Id;$r.original_manager_pid=$op.Id;$r.new_creation=$np.StartTime.ToUniversalTime().ToString('o')
}catch{$r.result='failed';$r.error=$_.Exception.Message}
$r|ConvertTo-Json -Depth 6|Set-Content (Join-Path $o 'result.json')
