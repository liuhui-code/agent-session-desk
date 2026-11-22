$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';Set-Location $b
$common=@('--backend',"$b\backend\psmux.exe",'--runtime-dir',"$b\desktop-runtime",'--namespace','asd-desktop-20261122')
$cli='$ErrorActionPreference="Stop";[Console]::InputEncoding=[Text.UTF8Encoding]::new();[Console]::OutputEncoding=[Text.UTF8Encoding]::new();function Record($o){[IO.File]::AppendAllText($env:ASD_TEST_JOURNAL,($o|ConvertTo-Json -Compress)+"`n",[Text.UTF8Encoding]::new($false))};Record @{kind="started";pid=$PID;inputEncoding=[Console]::InputEncoding.WebName;outputEncoding=[Console]::OutputEncoding.WebName};Write-Output "ASD UTF8 fixture ready";while($null -ne ($line=[Console]::ReadLine())){if($line -eq "exit"){break};Record @{kind="input";line=$line};Write-Output ("ASD UTF8 received: "+$line)}'
$enc=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($cli))
$asd="$b\product\asd-windows-amd64.exe"
$s=& $asd @common start --name 'UTF8 粘贴 D' --env ('ASD_TEST_JOURNAL='+"$b\desktop\fixture-D.jsonl") --cwd $b -- "$b\pwsh7\pwsh.exe" -NoLogo -NoProfile -EncodedCommand $enc
if($LASTEXITCODE -ne 0){throw 'UTF8 fixture create failed'}
$all=@(Get-Content desktop/created.json -Raw | ConvertFrom-Json);$all+=,($s|ConvertFrom-Json);$all|ConvertTo-Json -Depth 8|Set-Content desktop/created.json -Encoding UTF8
$s
