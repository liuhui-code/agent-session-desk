$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$p=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e';$o=Join-Path $b 'desk-audit-44aa154';$delivery=Join-Path $b 'desk-delivery-44aa154'
if(Test-Path $o){throw 'Fresh audit required'};New-Item -ItemType Directory $o|Out-Null
$r=@{result='started';product_launches=0;terminal_input=0;clipboard_apis=0;forced=0}
try{
& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $b 'desktop-api.ps1'))))
$h=Get-Content -Raw (Join-Path $p 'local-host.json')|ConvertFrom-Json;$m=Get-Content -Raw (Join-Path $p 'alpha.json')|ConvertFrom-Json
$new=Join-Path $p 'asd-desk-44aa154.exe';$expected='7e45eaf62b813d4b0c6e3943e29a146430c0202d6f68690df243fa00a01f27e7'
if($m.entrypoint-cne 'asd-desk-44aa154.exe'-or (Get-FileHash $new).Hash.ToLowerInvariant()-cne $expected){throw 'Product identity changed'}
$matches=@(Get-Process|Where-Object{$_.Path-ceq $new});if($matches.Count-ne 1-or $matches[0].Id-ne 21436){throw 'Manager identity changed'};$manager=$matches[0]
if($manager.StartTime.ToUniversalTime().ToString('o')-cne '2026-11-28T01:52:55.0438966Z'){throw 'Creation changed'}
if(@(Get-Process|Where-Object{$_.Path-ceq (Join-Path $p 'asd-five-5f5e6d3.exe')}).Count){throw 'Old manager remains'}
$pets=@([DeskFlow]::Find('ASD Blocking Reminders',$manager.Id));if($pets.Count-ne 1){throw 'Pet binding not unique'}
$owned=@();foreach($w in @([DeskFlow]::HostWindows($h.terminal.path))){[uint32]$hp=0;$null=[DeskFlow]::GetWindowThreadProcessId($w,[ref]$hp);$parent=$manager.Id;for($i=0;$i-lt 16-and $parent-gt 0;$i++){if($parent-eq $hp){$owned+=$w;break};$q=Get-CimInstance Win32_Process -Filter ('ProcessId='+$parent) -Property ParentProcessId;if(!$q){break};$parent=$q.ParentProcessId}}
if($owned.Count-ne 1-or [DeskFlow]::OwnerBindings($owned[0])-ne 1){throw 'Manager host not unique'}
$sets=@();foreach($name in @('live-before-private.json','live-after-private.json')){$rows=Get-Content -Raw (Join-Path $delivery $name)|ConvertFrom-Json;foreach($row in $rows){if($row.s-isnot [string]-or $row.r-isnot [string]-or $row.b-isnot [string]-or $row.s-cnotmatch '^[0-9a-f]{32}$'-or $row.r-cnotmatch '^[0-9a-f]{32}$'-or $row.b-cne ('asd_'+$row.s+'_'+$row.r)){throw 'Invalid identity projection'}};$sets+=,(ConvertTo-Json -InputObject @($rows|Sort-Object s,r) -Compress);$r.observed_identity_count=@($rows).Count};if($sets[0]-cne $sets[1]){throw 'Before/after identity mismatch'}
if([IO.File]::ReadAllText((Join-Path $delivery 'configs-before-private.txt'))-cne [IO.File]::ReadAllText((Join-Path $delivery 'configs-after-private.txt'))){throw 'Configuration mismatch'}
$r.result='updated-workbench-verified';$r.bound_host_count=1;$r.pet_count=1;$r.identity_projection_valid=$true;$r.before_after_equal=$true;$r.config_hashes_equal=$true;$r.manager_pid=$manager.Id;$r.working_bytes=$manager.WorkingSet64;$r.private_bytes=$manager.PrivateMemorySize64;$r.memory_sample_utc=[DateTime]::UtcNow.ToString('o');$r.hash=$expected
}catch{$r.result='failed';$r.error=$_.Exception.Message}
$r|ConvertTo-Json -Depth 5|Set-Content (Join-Path $o 'result.json')
