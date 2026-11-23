$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$registered=@();$registry=@()
function Get-AsdMetricPathKey {
    param([Parameter(Mandatory = $true)][string]$Path)
    # A lexical Windows path key also works in Linux-only synthetic graph tests.
    # The live collector requires full paths and compares the actual image path.
    # CIM can report \\?\C:\... while Process.Path reports C:\... for conhost.
    # Remove only an unambiguous extended DOS prefix. Extended paths otherwise
    # bypass Win32 parsing, so dot segments, device names, ADS, trailing dots or
    # spaces, empty components and forward slashes must not be normalized away.
    # https://learn.microsoft.com/en-us/windows/win32/fileio/naming-a-file
    # https://learn.microsoft.com/en-us/dotnet/standard/io/file-path-formats
    if ($Path.StartsWith('\\?\', [StringComparison]::Ordinal)) {
        $unchanged = $Path.ToLowerInvariant()
        if ($Path -notmatch '^\\\\\?\\[A-Za-z]:\\') { return $unchanged }
        $dosPath = $Path.Substring(4)
        foreach ($component in $dosPath.Substring(3).Split('\')) {
            if ($component.Length -eq 0 -or $component -in @('.', '..') -or
                $component.EndsWith('.') -or $component.EndsWith(' ') -or
                $component -match '[<>:"/|?*\x00-\x1F]') {
                return $unchanged
            }
            # Reserved DOS names remain reserved with an extension; include
            # Windows' three superscript-digit spellings and console devices.
            $baseName = $component.Split('.')[0].TrimEnd(' ')
            if ($baseName -match '^(CON|PRN|AUX|NUL|CONIN\$|CONOUT\$|COM[1-9\u00B9\u00B2\u00B3]|LPT[1-9\u00B9\u00B2\u00B3])$') {
                return $unchanged
            }
        }
        return $dosPath.ToLowerInvariant()
    }
    return $Path.Replace('/', '\').ToLowerInvariant()
}

function ReadShared([string]$Path){$f=[IO.FileStream]::new($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete));try{$r=[IO.StreamReader]::new($f);try{return $r.ReadToEnd()}finally{$r.Dispose()}}finally{$f.Dispose()}}
function Ended([long]$ProcessId,[long]$Filetime){$p=Get-Process -Id $ProcessId -ErrorAction SilentlyContinue;if(!$p){return $true};try{return ($p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $Filetime)}finally{$p.Dispose()}}
foreach($n in 1..5){$name=if($n -eq 1){'session-order-console-20261124'}else{'session-order-console-run'+$n+'-20261124'};$root=Join-Path $base $name;$ids=[IO.File]::ReadAllText((Join-Path $root 'identities.json'))|ConvertFrom-Json
foreach($id in $ids){$registered+=,@{attempt=$n;kind=$id.kind;pid=$id.pid;filetime=$id.filetime;original_ended=(Ended $id.pid $id.filetime)}}
foreach($file in @(Get-ChildItem (Join-Path $root 'driver') -Recurse -Filter '*.pid')){$text=(ReadShared $file.FullName).Trim();if($text -notmatch '^([1-9][0-9]*):([0-9]{17,20})$'){throw 'Driver identity syntax differs'};$processId=[long]$Matches[1];$filetime=[long]$Matches[2];$registry+=,@{attempt=$n;pid=$processId;filetime=$filetime;original_ended=(Ended $processId $filetime)}}
}
$images=@();foreach($name in @('asd-session-order.exe','session-order-session.exe','session-order-store.exe','session-order-manager.exe','session-order-ui.exe')){$path=Join-Path $base $name;$images+=,@{path=$path;remaining=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -and (Get-AsdMetricPathKey $_.ExecutablePath) -ceq (Get-AsdMetricPathKey $path)}).Count}}
$shared=@();foreach($name in @('psmux.exe','OpenConsole.exe')){$path=Join-Path $base ('terminal-bundle-20261123/bundle/'+$name);$shared+=,@{path=$path;remaining=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -and (Get-AsdMetricPathKey $_.ExecutablePath) -ceq (Get-AsdMetricPathKey $path)}).Count}}

$r=@{utc=[DateTime]::UtcNow.ToString('o');read_only=$true;shared_candidate_images=$shared;registered=$registered;registry=$registry;images=$images;registered_remaining=@($registered|Where-Object{!$_.original_ended}).Count;registry_remaining=@($registry|Where-Object{!$_.original_ended}).Count;remaining_exact_images=($images|ForEach-Object{$_.remaining}|Measure-Object -Sum).Sum}
$r|ConvertTo-Json -Depth 8|Set-Content (Join-Path $base 'session-order-final-audit.json') -Encoding utf8;$r|ConvertTo-Json -Depth 8
if($r.registered_remaining -or $r.registry_remaining -or $r.remaining_exact_images){throw 'Owned identities remain'}
