$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$root=Join-Path $env:USERPROFILE 'asd-acceptance-20261122/copy-final-20261123';$ns='asd-copy-final'
$sessions=[IO.File]::ReadAllText((Join-Path $root 'sessions.json'))|ConvertFrom-Json;$s=@($sessions|Where-Object{$_.title -eq 'Copy A'})
if($s.Count -ne 1){throw 'Expected synthetic Copy A identity'};$s=$s[0]
$data=Join-Path (Join-Path $root 'runtime') $ns;$reg=$ns+'__'+$s.backend_id
$port=[int]([IO.File]::ReadAllText((Join-Path $data ($reg+'.port')))).Trim();$key=([IO.File]::ReadAllText((Join-Path $data ($reg+'.key')))).Trim()
$tcp=[Net.Sockets.TcpClient]::new()
try{
 $tcp.Connect('127.0.0.1',$port);$stream=$tcp.GetStream();$stream.ReadTimeout=5000;$stream.WriteTimeout=3000
 $reader=[IO.StreamReader]::new($stream,[Text.UTF8Encoding]::new($false),$false,4096,$true)
 $bytes=[Text.Encoding]::UTF8.GetBytes("AUTH $key`ndump-state`n");$stream.Write($bytes,0,$bytes.Length);$stream.Flush();if($reader.ReadLine() -ne 'OK'){throw 'Authentication failed'}
 $line=$reader.ReadLine();if(!$line -or !$line.StartsWith('{')){throw 'No JSON metadata response'};$f=$line|ConvertFrom-Json
 if($f.session_name -ne $s.backend_id -or $f.layout.type -ne 'leaf'){throw 'Synthetic server identity mismatch'}
 $p=$f.layout
 @{session_id=$s.session_id;run_id=$s.run_id;pane=$p.id;cols=$p.cols;rows=$p.rows;copy_mode=$p.copy_mode;copy_cursor_row=$p.copy_cursor_row;copy_cursor_col=$p.copy_cursor_col;status_label=$f.status_label;status_cursor=$f.status_cursor;status_bytes=[Text.Encoding]::UTF8.GetByteCount([string]$f.status_message);selection_start_row=$p.sel_start_row;selection_end_row=$p.sel_end_row;highlight_count=@($p.copy_hl).Count;probe='authenticated dump-state only; no client attach, size or input; keys not logged'}|ConvertTo-Json|Set-Content (Join-Path $root 'copy-metadata-probe.json') -Encoding utf8
 Get-Content (Join-Path $root 'copy-metadata-probe.json') -Raw -Encoding utf8
}finally{$tcp.Dispose()}
