$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'tabs-ui-20261123'
$frames=[Collections.Generic.List[object]]::new();$tcp=$null
function Send([string]$line){$bytes=[Text.Encoding]::UTF8.GetBytes($line+"`n");$stream.Write($bytes,0,$bytes.Length);$stream.Flush()}
function Frame([string]$label){
 for($i=0;$i -lt 60;$i++){
  $line=$reader.ReadLine();if($null -eq $line){throw 'EOF'}
  if($line.StartsWith('{')){$f=$line|ConvertFrom-Json;$frames.Add(@{label=$label;frame=$f});return $f}
  if($line -ne 'NC' -and $line.Trim() -ne ''){throw 'Unexpected protocol response'}
 };throw 'No frame'
}
try{
 $all=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json
 $s=@($all|Where-Object{$_.title -eq 'Shell B'})[0]
 $data=Join-Path $root 'runtime/asd-tabs-ui';$reg="asd-tabs-ui__$($s.backend_id)"
 $port=[int]([IO.File]::ReadAllText((Join-Path $data "$reg.port"))).Trim()
 $key=([IO.File]::ReadAllText((Join-Path $data "$reg.key"))).Trim()
 $tcp=[Net.Sockets.TcpClient]::new();$tcp.Connect('127.0.0.1',$port)
 $stream=$tcp.GetStream();$stream.ReadTimeout=6000;$stream.WriteTimeout=3000
 $reader=[IO.StreamReader]::new($stream,[Text.UTF8Encoding]::new($false),$false,4096,$true)
 Send "AUTH $key";if($reader.ReadLine() -ne 'OK'){throw 'Auth failed'}
 Send 'PERSISTENT';Send 'client-attach'
 Send 'client-size 94 25';Send 'dump-state';$f=Frame 'initial'
 $cmd="1..30 | ForEach-Object { Write-Output ('W'*120) }"
 $b64=[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($cmd))
 Send "send-paste $b64 -t %$($f.layout.id)";Send "send-keys -t %$($f.layout.id) Enter"

 for($round=0;$round -lt 30;$round++){
 foreach($size in @(@(54,9),@(94,25),@(73,25),@(94,25))){
  Send "client-size $($size[0]) $($size[1])";Send 'dump-state'
  for($i=0;$i -lt 12;$i++){$f=Frame "$round-$($size[0])x$($size[1])";if($f.layout.cols -eq $size[0] -and $f.layout.rows -eq $size[1]){break}}
  Start-Sleep -Milliseconds 80
 }
 }
}finally{
 if($tcp){$tcp.Dispose()}
 $frames|ConvertTo-Json -Depth 18|Set-Content (Join-Path $root 'resize-probe-frames-run4.json') -Encoding utf8
 $bad=@($frames|Where-Object{$_.frame.layout.cursor_col -ge $_.frame.layout.cols -or $_.frame.layout.cursor_row -ge $_.frame.layout.rows -or $_.frame.layout.rows_v2.Count -gt $_.frame.layout.rows})
 $bad|ForEach-Object{@{label=$_.label;cols=$_.frame.layout.cols;rows=$_.frame.layout.rows;cursor_col=$_.frame.layout.cursor_col;cursor_row=$_.frame.layout.cursor_row;runrows=$_.frame.layout.rows_v2.Count}}|ConvertTo-Json
 @{total_frames=$frames.Count;out_of_bound_frames=$bad.Count}|ConvertTo-Json
}
