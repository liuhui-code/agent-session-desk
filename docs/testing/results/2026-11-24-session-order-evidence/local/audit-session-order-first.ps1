$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$registered=@();$registry=@()
function ReadShared([string]$Path){$f=[IO.FileStream]::new($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete));try{$r=[IO.StreamReader]::new($f);try{return $r.ReadToEnd()}finally{$r.Dispose()}}finally{$f.Dispose()}}
function Ended([long]$ProcessId,[long]$Filetime){$p=Get-Process -Id $ProcessId -ErrorAction SilentlyContinue;if(!$p){return $true};try{return ($p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $Filetime)}finally{$p.Dispose()}}
foreach($n in 1..3){$name=if($n -eq 1){'session-order-console-20261124'}else{'session-order-console-run'+$n+'-20261124'};$root=Join-Path $base $name;$ids=[IO.File]::ReadAllText((Join-Path $root 'identities.json'))|ConvertFrom-Json
foreach($id in $ids){$registered+=,@{attempt=$n;kind=$id.kind;pid=$id.pid;filetime=$id.filetime;original_ended=(Ended $id.pid $id.filetime)}}
foreach($file in @(Get-ChildItem (Join-Path $root 'driver') -Recurse -Filter '*.pid')){$text=(ReadShared $file.FullName).Trim();if($text -notmatch '^([1-9][0-9]*):([0-9]{17,20})$'){throw 'Driver identity syntax differs'};$processId=[long]$Matches[1];$filetime=[long]$Matches[2];$registry+=,@{attempt=$n;pid=$processId;filetime=$filetime;original_ended=(Ended $processId $filetime)}}
}
$images=@();foreach($name in @('asd-session-order.exe','session-order-session.exe','session-order-store.exe','session-order-manager.exe','session-order-ui.exe')){$path=Join-Path $base $name;$images+=,@{path=$path;remaining=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -eq $path}).Count}}
$r=@{utc=[DateTime]::UtcNow.ToString('o');read_only=$true;registered=$registered;registry=$registry;images=$images;registered_remaining=@($registered|Where-Object{!$_.original_ended}).Count;registry_remaining=@($registry|Where-Object{!$_.original_ended}).Count;remaining_exact_images=($images|ForEach-Object{$_.remaining}|Measure-Object -Sum).Sum}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $base 'session-order-final-audit.json') -Encoding utf8;$r|ConvertTo-Json -Depth 8
if($r.registered_remaining -or $r.registry_remaining -or $r.remaining_exact_images){throw 'Owned identities remain'}
