$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$result=@();$images=@()
foreach($n in 1..6){$name=if($n -eq 1){'long-field-console-20261124'}else{'long-field-console-run'+$n+'-20261124'};$root=Join-Path $base $name
foreach($file in @('manager-identity.json','shell-identity.json','driver-host.json')){$v=[IO.File]::ReadAllText((Join-Path $root $file))|ConvertFrom-Json;$p=Get-Process -Id ([int]$v.pid) -ErrorAction SilentlyContinue;$alive=$false;$currentTime=$null
if($p){try{$currentTime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();$alive=([long]$currentTime -eq [long]$v.filetime)}finally{$p.Dispose()}}
$result+=,@{attempt=$n;kind=$file;pid=$v.pid;filetime=$v.filetime;current_filetime=$currentTime;original_alive=$alive}
}
}
foreach($name in @('asd-long-field.exe','asd-long-field-final.exe','long-field-ui.exe','long-field-ui-final.exe')){$path=Join-Path $base $name;$images+=,@{path=$path;remaining=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $path}).Count}}
$r=@{utc=[DateTime]::UtcNow.ToString('o');originals=$result;images=$images;original_alive=@($result|Where-Object original_alive).Count;remaining_exact_images=($images|Measure-Object -Property remaining -Sum).Sum;read_only=$true}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $base 'long-field-final-audit.json') -Encoding utf8;$r|ConvertTo-Json -Depth 8
if($r.original_alive -or $r.remaining_exact_images){throw 'Original identities or exact images remain'}
