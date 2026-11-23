$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$found=@()
foreach($f in Get-ChildItem $base -Recurse -File -Filter 'conpty.dll'){
 $h=(Get-FileHash $f.FullName).Hash.ToLowerInvariant();$exe=Join-Path $f.Directory.FullName 'OpenConsole.exe';$eh=$null;if(Test-Path $exe){$eh=(Get-FileHash $exe).Hash.ToLowerInvariant()}
 $found+=,@{directory=[string]$f.Directory.FullName;dll_sha256=$h;exe_sha256=$eh;matches_pin=($h -eq 'e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492' -and $eh -eq 'b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775')}
}
$found|ConvertTo-Json -Depth 5
