$ErrorActionPreference='Stop'
$p=Get-Process -Id $PID
@{pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content alt-started.json -Encoding utf8
[Console]::Write("`e[?1049h`e[2J`e[HVT_ALT_READY 中文 α")
try{
 $line=[Console]::ReadLine()
 $cols=[Console]::WindowWidth;$rows=[Console]::WindowHeight
 @{pid=$PID;text=$line;columns=$cols;rows=$rows}|ConvertTo-Json|Set-Content alt-observed.json -Encoding utf8
}finally{[Console]::Write("`e[?1049l")}
'VT_ALT_DONE'
