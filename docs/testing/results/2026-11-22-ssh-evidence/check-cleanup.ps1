$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
Set-Location (Join-Path $env:USERPROFILE 'asd-acceptance-20261122')
$asd=(Resolve-Path 'product/asd-windows-amd64.exe').Path
$psmux=(Resolve-Path 'backend/psmux.exe').Path
$common=@('--backend',$psmux,'--runtime-dir',(Join-Path $PWD 'runtime'),'--namespace','asd-ssh-20261122')
function InvokeAsd([string[]]$Options){$out=& $asd @common @Options;if($LASTEXITCODE -ne 0){throw "asd exit $LASTEXITCODE"};return ($out -join "`n")}
$created=Get-Content created.json -Raw | ConvertFrom-Json
$list=InvokeAsd @('list','--json') | ConvertFrom-Json
$previews=@();$stops=@()
try {
 foreach($c in $created){$live=@($list | Where-Object {$_.session_id -eq $c.session_id});if($live.Count -ne 1 -or $live[0].run_id -ne $c.run_id){throw 'Run changed after UI reopen'};$previews+=,(InvokeAsd @('preview',$c.session_id) | ConvertFrom-Json)}
 if($previews[0].snapshot.text -notmatch 'ASD received: ASD_A_20261122' -or $previews[0].snapshot.text -match 'ASD_B_20261122'){throw 'A input isolation failure'}
 if($previews[1].snapshot.text -notmatch 'ASD received: ASD_B_20261122 中文' -or $previews[1].snapshot.text -match 'ASD_A_20261122'){throw 'B input isolation failure'}
} finally {foreach($c in $created){$stops+=,(InvokeAsd @('stop',$c.session_id) | ConvertFrom-Json)}}
$after=InvokeAsd @('list','--json') | ConvertFrom-Json
if(@($after).Count -ne 0){throw 'Private namespace not empty after stop'}
[ordered]@{utc=[DateTime]::UtcNow.ToString('o');before_stop=$list;previews=$previews;stop_results=$stops;after_stop=$after;identity_unchanged=$true;input_isolation=$true;scope='SSH PTY only; no Windows Terminal desktop or IME tested'} | ConvertTo-Json -Depth 8 | Set-Content check-cleanup-result.json -Encoding UTF8
Get-Content check-cleanup-result.json -Raw
