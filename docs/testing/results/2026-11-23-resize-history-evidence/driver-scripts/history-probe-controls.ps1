param([string]$Case='repro1',[switch]$NoResize,[switch]$NoCopy,[switch]$StartupOutput,[switch]$Direct,[switch]$IdleWriter)
$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base ('history-'+$Case)
if(Test-Path $root){throw 'Fresh root required'}
New-Item -ItemType Directory $root|Out-Null
$product=Join-Path $base 'copy-final2-product.exe';$backend=Join-Path $base 'backend/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$hashes=@{}
foreach($pair in @(@($product,'a2518bc0c28c4ff23ac544b440fe1fcc704c1d0e682bbfaf3cb7da91d65934b0'),@($backend,'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'),@($shell,'f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336'))){
 $h=(Get-FileHash $pair[0]).Hash.ToLowerInvariant();if($h -ne $pair[1]){throw 'Binary hash mismatch'};$hashes[[IO.Path]::GetFileName($pair[0])]=$h
}
$ns='asd-history-'+$Case;$runtime=Join-Path $root 'runtime';$data=Join-Path $runtime $ns
$g=@('--backend',$backend,'--runtime-dir',$runtime,'--namespace',$ns,'--state-dir',(Join-Path $root 'state'))
$fixture=Join-Path $root 'history-fixture.ps1'
@'
function global:prompt {'> '}
$p=Get-Process -Id $PID
@{pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content identity.json -Encoding utf8
'BOOT_READY'
while(!(Test-Path gate.txt)){Start-Sleep -Milliseconds 20}
Clear-Host
'SCROLL_NEEDLE_中文_α'
1..40|ForEach-Object {'history line '+$_}
'HISTORY_END'
'done'|Set-Content done.txt
if(Test-Path idle-writer.txt){while($true){Start-Sleep -Milliseconds 100}}
'@|Set-Content $fixture -Encoding utf8
$s=$null;$tcp=$null;$identity=$null;$pane=$null;$failure=$null
$stages=[Collections.Generic.List[object]]::new();$captures=[Collections.Generic.List[object]]::new()
$result=@{case=$Case;no_resize=[bool]$NoResize;no_copy=[bool]$NoCopy;hashes=$hashes;utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString;result='started'}
function Send([string]$line){$b=[Text.Encoding]::UTF8.GetBytes($line+"`n");$stream.Write($b,0,$b.Length);$stream.Flush()}
function Hex([string]$text){$h=([Text.Encoding]::UTF8.GetBytes($text)|ForEach-Object {$_.ToString('x2')}) -join ' ';Send "send-keys -t %$pane -H $h"}
function Lines($f){return @($f.layout.rows_v2|ForEach-Object {($_.runs.text -join '').TrimEnd()})}
function WaitFrame([scriptblock]$predicate){
 $until=[DateTime]::UtcNow.AddSeconds(12)
 while([DateTime]::UtcNow -lt $until){
  Send 'dump-state';$line=$reader.ReadLine();if($null -eq $line){throw 'EOF'}
  if(!$line.StartsWith('{')){if($line -eq 'NC' -or $line.Trim() -eq ''){continue};throw 'Unexpected non-JSON response'}
  $f=$line|ConvertFrom-Json
  if($f.session_name -ne $s.backend_id -or $f.layout.type -ne 'leaf' -or ($null -ne $pane -and $f.layout.id -ne $pane)){throw 'Session/pane changed'}
  if(& $predicate $f){return $f}
  Start-Sleep -Milliseconds 30
 }
 throw 'Expected frame not observed within 12 seconds'
}
function SaveStage([string]$label,$f){
 $p=$f.layout;$rows=@(Lines $f);$content=@($p.content|ForEach-Object {($_.text -join '').TrimEnd()})
 $projection=[ordered]@{label=$label;cols=$p.cols;rows=$p.rows;copy_mode=$p.copy_mode;scroll_offset=$p.scroll_offset;view_offset=$p.view_offset;copy_cursor_row=$p.copy_cursor_row;copy_cursor_col=$p.copy_cursor_col;status_message=$f.status_message;rows_v2_text=$rows;content_text=$content;rows_v2=$p.rows_v2;content=$p.content}
 $stages.Add($projection)
 Write-Host ($label+': '+$p.cols+'x'+$p.rows+' copy='+$p.copy_mode+' '+(($rows|Where-Object{$_ -match '^history line '}) -join ' | '))
}
function Capture([string]$label){
 if($Direct){$text=(& $backend -L $ns capture-pane -t ($s.backend_id+':.%'+$pane) -p -S -60|Out-String);if($LASTEXITCODE -ne 0){throw 'Direct capture failed'};$captures.Add(@{label=$label;snapshot=@{text=$text;truncated='unknown'}});return}
 $out=(& $product @g preview $s.session_id|Out-String);if($LASTEXITCODE -ne 0){throw 'Preview failed'}
 $v=$out|ConvertFrom-Json
 if($v.session_id -ne $s.session_id -or $v.run_id -ne $s.run_id){throw 'Preview identity changed'}
 $captures.Add(@{label=$label;snapshot=$v.snapshot})
}
try{
 if($IdleWriter){[IO.File]::WriteAllText((Join-Path $root 'idle-writer.txt'),'idle')}
 if($Direct){
  $env:PSMUX_NO_WARM='1';$env:PSMUX_ALLOW_NESTING='1';$env:PSMUX_DATA_DIR=$data;New-Item -ItemType Directory -Path $data|Out-Null
  $config=Join-Path $root 'psmux.conf';[IO.File]::WriteAllText($config,"set -g warm off`nset -g status off`n");$env:PSMUX_CONFIG_FILE=$config
  $backendID='history-direct-'+$Case;$reg=$ns+'__'+$backendID
  & $backend -L $ns new-session -d -s $backendID -c $root -x 120 -y 32 -- $shell -NoLogo -NoProfile -NoExit -File $fixture|Out-Null
  if($LASTEXITCODE -ne 0){throw 'Direct start failed'}
  $s=[pscustomobject]@{session_id='direct-'+$Case;run_id='direct-'+$Case;backend_id=$backendID}
 }else{
 $s=(& $product @g start --name 'Synthetic history resize' --cwd $root -- $shell -NoLogo -NoProfile -NoExit -File $fixture|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0){throw 'Start failed'}
 };$result.session=$s
 if($StartupOutput){[IO.File]::WriteAllText((Join-Path $root 'gate.txt'),'go');$until=[DateTime]::UtcNow.AddSeconds(10);while(!(Test-Path (Join-Path $root 'done.txt')) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 50};if(!(Test-Path (Join-Path $root 'done.txt'))){throw 'Startup output not complete'}}
 $reg=$ns+'__'+$s.backend_id
 $port=[int]([IO.File]::ReadAllText((Join-Path $data ($reg+'.port')))).Trim();$key=([IO.File]::ReadAllText((Join-Path $data ($reg+'.key')))).Trim()
 $tcp=[Net.Sockets.TcpClient]::new();$tcp.Connect('127.0.0.1',$port)
 $stream=$tcp.GetStream();$stream.ReadTimeout=15000;$stream.WriteTimeout=3000
 $reader=[IO.StreamReader]::new($stream,[Text.UTF8Encoding]::new($false),$false,4096,$true)
 Send "AUTH $key";if($reader.ReadLine() -ne 'OK'){throw 'Authentication failed'}
 Send 'PERSISTENT'
 if($StartupOutput){$f=WaitFrame {param($x) ((Lines $x) -join "`n") -match 'HISTORY_END'};$pane=$f.layout.id;SaveStage 'detached-before' $f;Capture 'detached-before'}
 Send 'client-attach';Send 'client-size 94 25'
 $f=WaitFrame {param($x) $x.layout.cols -eq 94 -and $x.layout.rows -eq 25 -and ($StartupOutput -or ((Lines $x) -join "`n") -match 'BOOT_READY')};$pane=$f.layout.id
 $identity=[IO.File]::ReadAllText((Join-Path $root 'identity.json'))|ConvertFrom-Json;$result.identity=$identity
 [IO.File]::WriteAllText((Join-Path $root 'gate.txt'),'go')
 $f=WaitFrame {param($x) ((Lines $x) -join "`n") -match 'HISTORY_END'};SaveStage 'live-before' $f;Capture 'live-before'
 if(!$NoCopy){
  Send "copy-mode -t %$pane"
  $f=WaitFrame {param($x) $x.layout.copy_mode};SaveStage 'copy-before' $f
  Hex ('?'+('x'*80)+'中文tail')
  $f=WaitFrame {param($x) $x.status_message -like '*中文tail*'};SaveStage 'prompt-before' $f
 }
 if(!$NoResize){
  Send 'client-size 61 18';$f=WaitFrame {param($x) $x.layout.cols -eq 61 -and $x.layout.rows -eq 18};SaveStage 'narrow' $f;Capture 'narrow'
 }
 if(!$NoCopy){Send "send-keys -t %$pane Escape";$f=WaitFrame {param($x) $x.status_message -notlike '*中文tail*'};SaveStage 'cancel-prompt' $f}
 if(!$NoResize){Send 'client-size 94 25';$f=WaitFrame {param($x) $x.layout.cols -eq 94 -and $x.layout.rows -eq 25};SaveStage 'wide-again' $f;Capture 'wide-again'}
 if($NoCopy){Send "copy-mode -t %$pane";$f=WaitFrame {param($x) $x.layout.copy_mode}}
 Hex '?SCROLL_NEEDLE_中文_α';$f=WaitFrame {param($x) $x.status_message -like '*SCROLL_NEEDLE*'};SaveStage 'needle-prompt' $f
 Send "send-keys -t %$pane Enter"
 $f=WaitFrame {param($x) @((Lines $x)|Where-Object{$_ -like '*SCROLL_NEEDLE*'}).Count -gt 0};SaveStage 'search-top' $f;Capture 'search-top'
 $ids=@((Lines $f)|ForEach-Object {if($_ -match '^history line (\d+)$'){[int]$Matches[1]}})
 $gaps=@();for($i=1;$i -lt $ids.Count;$i++){if($ids[$i] -ne $ids[$i-1]+1){$gaps+=,@($ids[$i-1],$ids[$i])}}
 $result.visible_ids=$ids;$result.gaps=$gaps;$result.result=if($gaps.Count){'reproduced-gap'}else{'no-gap-observed'}
} catch{$failure=$_.Exception.Message;$result.result='harness-failed';$result.error=$failure}
finally{
 if($tcp){$tcp.Dispose()}
 if($s){if($Direct){& $backend -L $ns kill-session -t $s.backend_id|Out-Null}else{& $product @g stop $s.session_id|Out-Null};if($LASTEXITCODE -ne 0){$result.cleanup_error='Stop failed'}}
 $listed=(& $product @g list --json|Out-String).Trim();$result.cleanup_json=$listed
 if($Direct){$connect=[Net.Sockets.TcpClient]::new();try{$connect.Connect('127.0.0.1',$port);$result.direct_endpoint_closed=$false}catch{$result.direct_endpoint_closed=$true}finally{$connect.Dispose()}}
 if($identity){$p=Get-Process -Id $identity.pid -ErrorAction SilentlyContinue;$result.original_ended=(!$p -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $identity.filetime -or $p.Path -ne $identity.path)}
 $stages.ToArray()|ConvertTo-Json -Depth 20|Set-Content (Join-Path $root 'stages.json') -Encoding utf8
 $captures.ToArray()|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'captures.json') -Encoding utf8
 $result|ConvertTo-Json -Depth 10|Set-Content (Join-Path $root 'result.json') -Encoding utf8
}
$result|ConvertTo-Json -Depth 10
if($failure -or $listed -ne '[]' -or $result.original_ended -ne $true -or $result.cleanup_error){throw 'Probe or cleanup failed; inspect result'}
