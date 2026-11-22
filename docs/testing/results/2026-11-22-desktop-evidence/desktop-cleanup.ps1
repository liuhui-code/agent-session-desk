$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';Set-Location $b
$asd="$b\product\asd-windows-amd64.exe";$common=@('--backend',"$b\backend\psmux.exe",'--runtime-dir',"$b\desktop-runtime",'--namespace','asd-desktop-20261122')
$record=Get-Content desktop/final-check.json -Raw|ConvertFrom-Json
$stops=@();foreach($c in $record.created){$raw=& $asd @common stop $c.session_id;if($LASTEXITCODE-ne 0){throw 'Precise session stop failed'};$stops+=,($raw|ConvertFrom-Json)}
$remaining=& $asd @common list --json;$result=$LASTEXITCODE;if($result-ne 0 -or (($remaining-join '')-replace '\s','')-ne'[]'){throw 'Namespace not empty'}
$processes=@();foreach($row in $record.fixture_processes){$p=Get-Process -Id $row.pid -ErrorAction SilentlyContinue;$same=$false;if($p -and $row.start_time){$same=$p.StartTime.ToUniversalTime().ToString('o')-eq$row.start_time};$processes+=@{fixture=$row.fixture;pid=$row.pid;original_identity_alive=$same}}
if(@($processes|Where-Object original_identity_alive).Count-ne 0){throw 'Owned fixture still alive'}
$terminal=(Get-ChildItem terminal -Recurse -Filter WindowsTerminal.exe | Select-Object -First 1).FullName
$settings=Join-Path (Split-Path $terminal) 'settings/settings.json';$config=Get-Content $settings -Raw|ConvertFrom-Json;$config.profiles.list[0].font.face='Consolas';$config.profiles.list[0].font.size=12;$config|ConvertTo-Json -Depth 10|Set-Content $settings -Encoding UTF8
@{utc=[DateTime]::UtcNow.ToString('o');stop_results=$stops;remaining=@();fixture_processes=$processes;private_font_restored=$true;scope='Only exact Session IDs created by this test; PID plus start-time verified'}|ConvertTo-Json -Depth 8|Set-Content desktop/cleanup.json -Encoding UTF8
Get-Content desktop/cleanup.json -Raw
