from pathlib import Path
import json
out=Path('/private/tmp/asd-session-stop-20261124'); hashes=json.loads((out/'binary-hashes.json').read_text())
s=Path('/private/tmp/asd-session-order-20261124/run-session-order-tests.ps1').read_text();a=s.index('$entries=');b=s.index('$results=',a)
entries=[dict(file=n,package=p,sha256=hashes[n]) for p,n in [('manager','session-stop-manager.exe'),('ui','session-stop-ui.exe')]]
s=s[:a]+"$entries=ConvertFrom-Json @'\n"+json.dumps(entries)+"\n'@\n"+s[b:]
(out/'run-session-stop-tests.ps1').write_text(s.replace('session-order-native-20261124','session-stop-native-20261124'))
s=Path('/private/tmp/asd-session-order-20261124/run-session-order-console.ps1').read_text();prefix=s[:s.index('try{\n$fixture=')]
prefix=prefix.replace('session-order-console-run5-20261124','session-stop-console-20261124').replace('asd-session-order.exe','asd-session-stop.exe').replace('asd-session-order-run5','asd-session-stop').replace('asd-session-order-driver-run5','asd-session-stop-driver').replace('order','stop')
prefix=prefix.replace('32e0d77538709ef269d8c7f743851df02c6bf7947b7346da960b0104f2038289',hashes['asd-session-stop.exe'])
prefix=prefix.replace(";$legacy=Join-Path $base 'asd-long-field-final.exe'",'').replace("@($legacy,'2f022aaddec1aa8fc247c06e9557a97a7e2af9a8d227e7dd3e47dca6719e10a8'),",'')
fixture=s[s.index('try{\n$fixture='):s.index("foreach($name in @('A','B')){$v=")].replace('ORDER_READY','STOP_READY')
body=r'''foreach($name in @('A','B')){$v=(& $product @g start --name ('Stop '+$name) --cwd $root -- $shell -NoLogo -NoProfile -NoExit -File $fixture $name|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0){throw 'Shell start failed'};$sessions+=,$v;$sessions|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'sessions-before.json') -Encoding utf8}
foreach($name in @('A','B')){$id=ReadFixtureIdentity (Join-Path $root ($name+'.identity.json'));$p=Get-Process -Id $id.pid;if($p.Path -ne $shell -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime){throw 'Original shell differs'};Register $p ('shell-'+$name);$shells+=,$p}
& $product @g history|Set-Content (Join-Path $root 'history-before.json') -Encoding utf8
StartUI 'ui1';Await 'STOP_READY A';Text "Write-Output ('STOP_A:'+`$PID)";Await "Write-Output ('STOP_A:'+`$PID)";Capture 'A-original-draft'
Key 'C-PageDown';Await 'STOP_READY B';Text "Write-Output ('STOP_B:'+`$PID)";Await "Write-Output ('STOP_B:'+`$PID)";Capture 'B-original-draft'
Key 'C-PageUp';Await "Write-Output ('STOP_A:'+`$PID)";Key 'F6';Await '管理模式';Key 'Down';Text 'd';Await '停止会话';Await $sessions[1].run_id;Capture 'B-selected-A-active-default-cancel'
Key 'Enter';Await '管理模式';Await "Write-Output ('STOP_A:'+`$PID)";if($shells[0].HasExited -or $shells[1].HasExited){throw 'Default cancel stopped shell'};Capture 'cancel-preserved-both'
Text 'd';Await '停止会话';$null=Mux @('resize-window','-t',$script:stopTarget,'-x','120','-y','10');Await '请放大窗口';Text 'y';Await '请放大窗口';if(@(Rows).Count -ne 2 -or $shells[1].HasExited){throw 'Unreadable review stopped target'};Capture 'narrow-review-refused'
$null=Mux @('resize-window','-t',$script:stopTarget,'-x','120','-y','32');Await $sessions[1].run_id;Capture 'B-full-review-restored';Text 'y'
$until=[DateTime]::UtcNow.AddSeconds(10);do{$rows=@(Rows);if($rows.Count -eq 1 -and $rows[0].session_id -eq $sessions[0].session_id){break};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until)
if($rows.Count -ne 1 -or $rows[0].run_id -ne $sessions[0].run_id -or !$shells[1].WaitForExit(10000) -or $shells[0].HasExited){throw 'Selected B stop did not preserve original A'}
SnapshotRows 'after-B-stop' $rows;Await '管理模式';Await "Write-Output ('STOP_A:'+`$PID)";Capture 'B-stopped-A-draft-preserved'
$history=(& $product @g history|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0){throw 'History query failed'};SnapshotRows 'history-after-B-stop' @($history)
$hB=@($history|Where-Object{$_.session.session_id -eq $sessions[1].session_id});$hA=@($history|Where-Object{$_.session.session_id -eq $sessions[0].session_id});if($hB.Count -ne 1 -or !$hB[0].stop_confirmed_at -or $hA.Count -ne 1 -or $hA[0].stop_confirmed_at){throw 'Stopped history confused with live A'}
Key 'C-PageDown';Await '目标终端当前不可交互';Capture 'B-view-retained';Text 'x';Await "Write-Output ('STOP_A:'+`$PID)";Capture 'B-view-closed-A-remains'
Text 'd';Await '停止会话';Await $sessions[0].run_id;Capture 'A-reviewed';Key 'Left';Key 'Enter'
$until=[DateTime]::UtcNow.AddSeconds(10);do{$rows=@(Rows);if($rows.Count -eq 0){break};Start-Sleep -Milliseconds 80}while([DateTime]::UtcNow -lt $until)
if(!$shells[0].WaitForExit(10000) -or $rows.Count -ne 0){throw 'Active A stop missing'};Capture 'A-stopped-view-retained'
$history=(& $product @g history|Out-String)|ConvertFrom-Json;if($LASTEXITCODE -ne 0 -or @($history).Count -ne 2 -or @($history|Where-Object{!$_.stop_confirmed_at}).Count -ne 0){throw 'Original stopped history missing'};SnapshotRows 'history-final' @($history)
StopUI
StartUI 'ui2';Await '暂无打开的 Tab';Await '管理模式';Capture 'reopened-no-live-sessions';Text 'h';Await '历史记录';Capture 'reopened-history-preserved';Key 'Esc';Await '管理模式';StopUI
$success=$true
}catch{Event 'failure' @{error=$_.Exception.Message};throw}finally{
$remaining=@(Rows);foreach($v in $sessions){$found=@($remaining|Where-Object{$_.session_id -eq $v.session_id -and $_.run_id -eq $v.run_id -and $_.backend_id -eq $v.backend_id});if($found.Count -eq 1){& $product @g stop $v.session_id|Set-Content (Join-Path $root ('cleanup-stop-'+$v.session_id+'.json')) -Encoding utf8}}
if($script:stopTarget -and $script:stopManagers.Count -gt 0 -and !$script:stopManagers[-1].HasExited){$forced++;$name=$script:stopTarget.Split(':')[0];try{$null=Mux @('kill-session','-t',$name)}catch{Event 'cleanup-error' @{error=$_.Exception.Message}}}
foreach($p in @($shells)+@($script:stopManagers)+@($script:stopServers)){$null=$p.WaitForExit(10000)}
$ended=@();foreach($id in $script:stopIdentities){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$alive=$false;if($p){$alive=$p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $id.filetime};$ended+=,@{pid=$id.pid;filetime=$id.filetime;kind=$id.kind;original_ended=(!$alive)}}
$r=@{utc=[DateTime]::UtcNow.ToString('o');result=$(if($success){'passed'}else{'failed'});captures=$script:stopCaptureNumber;events=$events.Count;product_sha256=(Get-FileHash $product).Hash;windows=[Environment]::OSVersion.VersionString;originals=$ended;forced_product_ends=$forced;clipboard_api_calls=0;desktop_surface='none: isolated ConPTY';remaining_exact_product=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $product}).Count}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'outcome.json') -Encoding utf8;$r|ConvertTo-Json -Depth 8
}
'''
(out/'run-session-stop-console.ps1').write_text(prefix+fixture+body)
