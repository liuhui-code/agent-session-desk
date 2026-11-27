$ErrorActionPreference='Stop';$PSNativeCommandUseErrorActionPreference=$false
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$o=Join-Path $b 'preset-cleanup3';New-Item -ItemType Directory $o|Out-Null
$r=@{result='started';stops=@();forced=0};$handles=@()
try{
$old=Join-Path $b 'asd-output.exe';$backend=Join-Path $b 'backend/psmux.exe';$fixture=Join-Path $b 'asd-fixture.exe'
foreach($v in @(@($old,'da88e3ac5622c817cabc44231ab81862a6db7c9af865de7702f96e1fa84acc7d'),@($backend,'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'),@($fixture,'f9a30f0dae7bd6a9fffe7deedf8a194405d67d651aad0b99cc20296e8cb5d9d6'))){if((Get-FileHash $v[0]).Hash.ToLowerInvariant()-cne $v[1]){throw 'Pinned old test artifact differs'}}
$root=Join-Path $b 'm1-final-ui-run1';$g=@('--backend',$backend,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-m1-final','--state-dir',(Join-Path $root 'state'))
function OldList{$t=& $old @g list --json 2>$null;if($LASTEXITCODE-ne 0){throw 'Old list failed'};$data=$t|Out-String|ConvertFrom-Json;foreach($record in $data){Write-Output $record}}
$rows=@(OldList);$expected=@{'87ab7ec59c89e69f879906ed119a6810'=@('5c54b2cabb6db2853d218d7b0302f40f','A 开发');'8601152e7047821cc3b749bd3b9c2413'=@('78d8aba6695804d9716f286c90b5d690','B 守护');'2d59e844285a86649248eb5753b37555'=@('27ffa2ea17d76bfddeb9ca16d9b27293','C 旁观')}
if($rows.Count-ne 3){throw 'Test scope changed'}
foreach($s in $rows){$v=$expected[[string]$s.session_id];if(!$v-or $s.run_id-cne $v[0]-or $s.title-cne $v[1]-or $s.command-ine $fixture-or $s.backend_id-cne ('asd_'+$s.session_id+'_'+$s.run_id)){throw 'Test target identity changed'}}
$p=Join-Path ([Environment]::GetFolderPath('Desktop')) 'Agent Session Desk Alpha 8234c4e';$m=Get-Content -Raw (Join-Path $p 'alpha.json')|ConvertFrom-Json;$daily=Join-Path $p $m.entrypoint
$d=@('--terminal-bundle',(Join-Path $p 'terminal-runtime'),'--runtime-dir',(Join-Path $p 'data/runtime'),'--namespace','asd-alpha-8234c4e','--state-dir',(Join-Path $p 'data/state'))
function DailyIds{$t=& $daily @d list --json 2>$null;if($LASTEXITCODE-ne 0){throw 'Daily observation failed'};$data=$t|Out-String|ConvertFrom-Json;$ids=@(foreach($entry in $data){[pscustomobject][ordered]@{s=$entry.session_id;r=$entry.run_id;b=$entry.backend_id}});return ConvertTo-Json -InputObject @($ids|Sort-Object s,r) -Compress}
$before=DailyIds
$env:PSMUX_DATA_DIR=Join-Path $root 'runtime/asd-m1-final';$env:PSMUX_CONFIG_FILE=Join-Path $env:PSMUX_DATA_DIR 'psmux-m0.conf';$env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1'
foreach($s in $rows){if($s.terminal_state-ne 'available'){continue};$text=& $backend '-L' 'asd-m1-final' '-f' $env:PSMUX_CONFIG_FILE 'display-message' '-p' '-t' $s.backend_id '#{pane_pid}' 2>$null;if($LASTEXITCODE-ne 0){throw 'Fixture PID observation failed'};[int]$fixtureId=0;if(![int]::TryParse(($text|Out-String).Trim(),[ref]$fixtureId)){throw 'Fixture PID invalid'};$process=Get-Process -Id $fixtureId -ErrorAction Stop;$null=$process.Handle;if($process.Path-ine $fixture){throw 'Original pane is not the pinned test fixture'};$handles+=$process}
$r.labels=@($rows|ForEach-Object{[string]$_.title});$r.fixture_handles=$handles.Count
foreach($s in $rows){$null=& $old @g stop $s.session_id 2>$null;$code=$LASTEXITCODE;$r.stops+=@{id=$s.session_id;code=$code};if($code-ne 0){throw 'Stop failed; not retried'}}
$r.remaining=@(OldList).Count;if($r.remaining-ne 0){throw 'Test scope still has sessions'}
foreach($process in $handles){if(!$process.WaitForExit(10000)){throw 'Fixture exit not observed; no force kill'}}
$r.fixture_exits_observed=$handles.Count;$r.daily_preserved=($before-ceq (DailyIds));if(!$r.daily_preserved){throw 'Daily set changed; no destructive recovery attempted'}
$r.result='cleaned'
}catch{$r.result='failed';$r.error=$_.Exception.Message}
$r|ConvertTo-Json -Depth 6|Set-Content -Encoding utf8 (Join-Path $o 'r')
