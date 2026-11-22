$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'history-tool-candidate1'
$r=[IO.File]::ReadAllText((Join-Path $root 'result.json'))|ConvertFrom-Json
$observed=[IO.File]::ReadAllText((Join-Path $root 'startup-inspection.json'))|ConvertFrom-Json
$identities=@($observed.matching_test_processes);if($identities.Count -ne 2){throw 'Expected registered server and fixture'}
foreach($i in $identities){$p=Get-Process -Id $i.pid;if($p.Path -ne $i.path -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $i.filetime){throw 'Original identity changed'}}
$data=Join-Path (Join-Path $root 'runtime') $r.namespace;$ports=@(Get-ChildItem $data -Filter '*.port')
if($ports.Count -ne 1){throw 'Expected one private endpoint'};$entry=$ports[0];$reg=$entry.BaseName
$expected=$reg.Substring(($r.namespace+'__').Length)
$port=[int]([IO.File]::ReadAllText($entry.FullName)).Trim();$key=([IO.File]::ReadAllText((Join-Path $data ($reg+'.key')))).Trim()
$tcp=[Net.Sockets.TcpClient]::new();$tcp.Connect('127.0.0.1',$port)
try{
 $stream=$tcp.GetStream();$stream.ReadTimeout=5000;$stream.WriteTimeout=3000;$reader=[IO.StreamReader]::new($stream,[Text.UTF8Encoding]::new($false),$false,4096,$true)
 function Send([string]$s){$b=[Text.Encoding]::UTF8.GetBytes($s+"`n");$stream.Write($b,0,$b.Length);$stream.Flush()}
 Send "AUTH $key";if($reader.ReadLine() -ne 'OK'){throw 'Authentication failed'};Send 'PERSISTENT';Send 'dump-state'
 $f=$reader.ReadLine()|ConvertFrom-Json
 if($f.session_name -ne $expected -or $f.layout.type -ne 'leaf'){throw 'Synthetic Session changed'}
 Send 'kill-session'
}finally{$tcp.Dispose()}
$until=[DateTime]::UtcNow.AddSeconds(15)
do{$remaining=@($identities|Where-Object{$p=Get-Process -Id $_.pid -ErrorAction SilentlyContinue;$p -and $p.Path -eq $_.path -and $p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $_.filetime});if(!$remaining.Count){break};Start-Sleep -Milliseconds 100}while([DateTime]::UtcNow -lt $until)
$a=@{result=$(if($remaining.Count -or (Test-Path $entry.FullName)){'failed'}else{'passed'});cleanup='authenticated original Session kill-session; no PID/name kill';original_identities=$identities;original_processes_remaining=$remaining.Count;port_removed=!(Test-Path $entry.FullName)}
$a|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'cleanup-audit.json') -Encoding utf8;$a|ConvertTo-Json -Depth 6
if($remaining.Count -or !$a.port_removed){throw 'Original candidate or fixture remains'}
$new=Join-Path $base 'history-backend-candidate';if(Test-Path $new){throw 'Fresh candidate directory required'};New-Item -ItemType Directory $new|Out-Null
Copy-Item (Join-Path $base 'psmux-history-candidate.exe') (Join-Path $new 'psmux.exe')
if((Get-FileHash (Join-Path $new 'psmux.exe')).Hash.ToLowerInvariant() -ne 'ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'){throw 'Candidate copy differs'}
