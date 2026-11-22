$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new();$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$out="$b\detach-diagnosis"
$s=Get-Content "$out\created.json" -Raw|ConvertFrom-Json;$before=Get-Content "$out\final-check.json" -Raw|ConvertFrom-Json
$asd="$b\asd-detach-guidance.exe";$o=@('--backend',"$b\backend\psmux.exe",'--runtime-dir',"$b\detach-runtime",'--namespace','asd-detach-20261122')
$stopped=& $asd @o stop $s.session_id;if($LASTEXITCODE-ne 0){throw 'Stop failed'}
$after=& $asd @o list --json;if($LASTEXITCODE-ne 0 -or ($after-join'').Trim()-ne'[]'){throw 'Diagnostic namespace not empty'}
$p=Get-Process -Id $before.fixture_pid -ErrorAction SilentlyContinue
$same=$false;if($p){$same=$p.StartTime.ToUniversalTime().ToString('o')-eq$before.fixture_start_time};if($same){throw 'Original fixture still alive'}
@{utc=[DateTime]::UtcNow.ToString('o');stop=($stopped|ConvertFrom-Json);after_raw=($after-join "`n");original_fixture_alive=$same}|ConvertTo-Json -Depth 6|Set-Content "$out\cleanup.json" -Encoding UTF8
Get-Content "$out\cleanup.json" -Raw
