$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'tabs-wire-probe-2'
if(Test-Path $root){throw 'Fresh root required'}
New-Item -ItemType Directory $root|Out-Null
$product=Join-Path $base 'asd-output.exe';$psmux=Join-Path $base 'backend/psmux.exe'
$hash=(Get-FileHash $psmux).Hash.ToLowerInvariant()
if($hash -ne 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'){throw 'Wrong psmux'}
$ns='asd-tabs-wire';$runtime=Join-Path $root 'runtime';$data=Join-Path $runtime $ns
$g=@('--backend',$psmux,'--runtime-dir',$runtime,'--namespace',$ns,'--state-dir',(Join-Path $root 'state'))
$s=$null;$tcp=$null;$frames=[Collections.Generic.List[object]]::new();$failure=$null
function Send([string]$line){$bytes=[Text.Encoding]::UTF8.GetBytes($line+"`n");$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}
function Frame([string]$label){
 for($i=0;$i -lt 60;$i++){
  $line=$reader.ReadLine();if($null -eq $line){throw 'EOF'}
  if($line.StartsWith('{')){$f=$line|ConvertFrom-Json;$frames.Add(@{label=$label;frame=$f});return $f}
  if($line -ne 'NC' -and $line.Trim() -ne ''){throw "Unexpected frame $line"}
 }
 throw 'No frame'
}
function Text($f){return (($f.layout.rows_v2|ForEach-Object {($_.runs.text -join '')}) -join "`n")}
try{
 $shell=Join-Path $base 'pwsh7/pwsh.exe'
 $s=(& $product @g start --name 'Tab shell wire probe' --cwd $root -- $shell -NoLogo -NoProfile|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0){throw 'Start failed'}
 $reg="$ns`__$($s.backend_id)"
 $port=[int]([IO.File]::ReadAllText((Join-Path $data "$reg.port"))).Trim()
 $key=([IO.File]::ReadAllText((Join-Path $data "$reg.key"))).Trim()
 $tcp=[Net.Sockets.TcpClient]::new();$tcp.Connect('127.0.0.1',$port)
 $stream=$tcp.GetStream();$stream.ReadTimeout=10000;$stream.WriteTimeout=3000
 $reader=[IO.StreamReader]::new($stream,[Text.UTF8Encoding]::new($false),$false,4096,$true)
 Send "AUTH $key";if($reader.ReadLine() -ne 'OK'){throw 'Auth failed'}
 Send 'PERSISTENT';Send 'client-attach';Send 'client-size 80 20';Send 'dump-state'
 $f=Frame 'initial'
 if($f.layout.type -ne 'leaf' -or $f.layout.cols -ne 80 -or $f.layout.rows -ne 20){throw 'Wrong leaf/size'}
 $pane=$f.layout.id
 $command="Write-Output 'ASD_TAB_WIRE_2'; 'executed' | Set-Content observed-2.txt`r"
 $hex=([Text.Encoding]::UTF8.GetBytes($command)|ForEach-Object {$_.ToString('x2')}) -join ' '
 Send "send-keys -t %$pane -H $hex"
 Send 'dump-state'
 $ok=$false
 for($i=0;$i -lt 40;$i++){$f=Frame 'echo';if((Text $f) -match 'ASD_TAB_WIRE_2'){$ok=$true;break}}
 if(!$ok){throw 'Echo not observed'}
 $until=[DateTime]::UtcNow.AddSeconds(10);$observed=Join-Path $root 'observed-2.txt'
 while(!(Test-Path $observed) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 50}
 if(!(Test-Path $observed) -or [IO.File]::ReadAllText($observed).Trim() -ne 'executed'){throw 'Shell execution not independently observed'}
 Send 'client-size 92 24';Send 'dump-state'
 $ok=$false
 for($i=0;$i -lt 40;$i++){$f=Frame 'resize';if($f.layout.cols -eq 92 -and $f.layout.rows -eq 24){$ok=$true;break}}
 if(!$ok){throw 'Resize not observed'}
 $tcp.Dispose();$tcp=$null
 $listed=(& $product @g list --json|Out-String)|ConvertFrom-Json
 if($listed.run_id -notcontains $s.run_id){throw 'Disconnect lost Run'}
 $result=@{result='passed';backend_hash=$hash;session=$s;wire_session=$f.session_name;pane=$pane;frames=$frames.Count;checks=@('native persistent auth','styled screen snapshot','explicit pane hex input with independent execution file','client resize','disconnect preserves run')}
} catch {$failure=$_.Exception.Message;$result=@{result='failed';error=$failure;frames=$frames.Count}}
finally{
 if($tcp){$tcp.Dispose()}
 if($s){& $product @g stop $s.session_id|Out-Null}
 $text=(& $product @g list --json|Out-String).Trim()
 $result['cleanup_json']=$text
 if($text -ne '[]'){throw 'Cleanup not empty'}
 $result['remaining_sessions']=0
 $frames|ConvertTo-Json -Depth 18|Set-Content (Join-Path $root 'frames.json') -Encoding utf8
 $result|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'result.json') -Encoding utf8
}
$result|ConvertTo-Json -Depth 8
if($failure){throw $failure}
