$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$p=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e';$o=Join-Path $b 'five-switch-audit-5f5e6d3'
if(Test-Path $o){throw 'Fresh audit required'};New-Item -ItemType Directory $o|Out-Null
$r=@{result='started';launches=0;terminal_input=0;clipboard_apis=0;forced=0;stages=@()}
function Stage([string]$name){$r.stages+=@{name=$name;at=[DateTime]::UtcNow.ToString('o')};$r|ConvertTo-Json -Depth 5|Set-Content (Join-Path $o 'progress.json')}
try{
Stage 'guards'
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $b 'desktop-api.ps1'))))
$m=Get-Content -Raw (Join-Path $p 'alpha.json')|ConvertFrom-Json;$h=Get-Content -Raw (Join-Path $p 'local-host.json')|ConvertFrom-Json
$new=Join-Path $p 'asd-five-5f5e6d3.exe';$expected='6a96f1c34aeb5c1e51bd2a171e46696e7f9e71ef803ecaae3d5ce99c3ca3c060'
if($m.entrypoint-cne 'asd-five-5f5e6d3.exe'-or (Get-FileHash $new).Hash.ToLowerInvariant()-cne $expected){throw 'Product identity changed'}
$matches=@(Get-Process -Name 'asd-five-5f5e6d3' -ErrorAction SilentlyContinue|Where-Object{$_.Path-ceq $new});if($matches.Count-ne 1){throw 'Expected one existing manager; no launch attempted'};$manager=$matches[0];$null=$manager.Handle
$cim=Get-CimInstance Win32_Process -Filter ('ProcessId='+$manager.Id) -Property CommandLine;if($cim.CommandLine-notmatch '--inside-terminal'){throw 'Existing process is not the inner workbench'}
$old=Join-Path $p 'asd-idle-1d842c0.exe';if(@(Get-Process -Name 'asd-idle-1d842c0' -ErrorAction SilentlyContinue|Where-Object{$_.Path-ceq $old}).Count){throw 'Old product still running'}
Stage 'window-bindings'
$pets=@([DeskFlow]::Find('ASD Blocking Reminders',$manager.Id));if($pets.Count-ne 1){throw 'Original manager pet not observable'}
$owned=@();foreach($w in @([DeskFlow]::HostWindows($h.terminal.path))){[uint32]$hp=0;$null=[DeskFlow]::GetWindowThreadProcessId($w,[ref]$hp);$parent=$manager.Id;for($i=0;$i-lt 16-and $parent-gt 0;$i++){if($parent-eq $hp){$owned+=$w;break};$q=Get-CimInstance Win32_Process -Filter ('ProcessId='+$parent) -Property ParentProcessId;if(!$q){break};$parent=$q.ParentProcessId}}
if($owned.Count-ne 1-or [DeskFlow]::OwnerBindings($owned[0])-ne 1){throw 'Manager host binding not unique'}
$r.manager_pid=$manager.Id;$r.manager_start=$manager.StartTime.ToUniversalTime().ToString('o');$r.pet_count=1;$r.bound_host_count=1
Stage 'metadata-and-live'
$g=@('--terminal-bundle',(Join-Path $p 'terminal-runtime'),'--runtime-dir',(Join-Path $p 'data/runtime'),'--namespace','asd-alpha-8234c4e','--state-dir',(Join-Path $p 'data/state'))
$backup=Join-Path $b 'five-switch-delivery-5f5e6d3/metadata-v11-private.sqlite3'
foreach($file in @($backup,(Join-Path $p 'data/state/metadata.sqlite3'))){$check=(& $new metadata check $file 2>$null|Out-String|ConvertFrom-Json);if($LASTEXITCODE-ne 0-or $check.schema_version-ne 11-or $check.integrity-cne 'ok'){throw 'Metadata validation failed'}}
$text=& $new @g list --json 2>$null;if($LASTEXITCODE-ne 0){throw 'Current live observation failed'};[IO.File]::WriteAllText((Join-Path $o 'live-private.json'),($text|Out-String),[Text.UTF8Encoding]::new($false))
$audit=Join-Path $b 'five-identity-audit.exe';if((Get-FileHash $audit).Hash.ToLowerInvariant()-cne '5e55f7368aaa2f20326611fa6009087841a657f2fc1a1b5fdb045e27c7c33bf5'){throw 'Audit image changed'}
$scope=('psmux:'+(Join-Path $p 'data/runtime/asd-alpha-8234c4e')).ToLowerInvariant();$e=Start-Process -FilePath $audit -ArgumentList ('"'+$backup+'"'),('"'+(Join-Path $p 'data/state/metadata.sqlite3')+'"'),('"'+(Join-Path $o 'live-private.json')+'"'),('"'+$scope+'"') -Wait -PassThru -NoNewWindow -RedirectStandardOutput (Join-Path $o 'identity.json') -RedirectStandardError (Join-Path $o 'identity-stderr.log');if($e.ExitCode-ne 0){throw 'Read-only metadata identity audit failed'}
$r.identities=Get-Content -Raw (Join-Path $o 'identity.json')|ConvertFrom-Json
if(!$r.identities.original_metadata_identities_preserved){throw 'Original metadata identities missing'}
$r.config_last_write_before_update=@();foreach($name in @('ui.toml','launch.toml','last-launch.toml','recent-dirs.toml','keybindings.toml')){$f=Join-Path $p ('data/state/'+$name);if(Test-Path $f){$item=Get-Item $f;$r.config_last_write_before_update+=@{name=$name;before=($item.LastWriteTimeUtc-lt [DateTime]::Parse('2026-11-27T15:42:55Z').ToUniversalTime())}}}
$r.metadata_schema=11;$r.metadata_integrity='ok';$r.hash=$expected;$r.source='5f5e6d3';$r.result='existing-updated-workbench-verified'
}catch{$r.result='failed';$r.error=$_.Exception.Message}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $o 'result.json')
