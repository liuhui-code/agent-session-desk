
$ErrorActionPreference='Stop'
$p=Get-Process -Id $PID
@{PID=$PID;Created=$p.StartTime.ToUniversalTime().ToFileTimeUtc();Path=$p.Path}|ConvertTo-Json|Set-Content root.json -Encoding utf8NoBOM
Clear-Host
1..40|ForEach-Object{'history line '+$_}
[IO.File]::WriteAllText((Join-Path $PWD 'seeded.txt'),'seeded')
'BUNDLE_INPUT_READY'
$line=[Console]::ReadLine()
[IO.File]::WriteAllText((Join-Path $PWD 'normal.txt'),$line,[Text.UTF8Encoding]::new($false))
while(!(Test-Path alt-gate.txt)){Start-Sleep -Milliseconds 20}
[Console]::Write("`e[?1049h`e[2J`e[HBUNDLE_ALT_READY")
try{
 $line=[Console]::ReadLine()
 [IO.File]::WriteAllText((Join-Path $PWD 'alt.txt'),$line,[Text.UTF8Encoding]::new($false))
 [IO.File]::WriteAllText((Join-Path $PWD 'size.txt'),([string][Console]::WindowWidth+','+[string][Console]::WindowHeight))
}finally{[Console]::Write("`e[?1049l")}
'BUNDLE_ALT_DONE'
