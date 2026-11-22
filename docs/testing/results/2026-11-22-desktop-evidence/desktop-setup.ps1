$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';Set-Location $base
foreach($x in @(@('Microsoft.WindowsTerminal_1.25.3131.0_x64.zip','273d6f1502eca1b5bd95c43f5d4d7ec72cc503df340102f8659717e5bead3437'),@('PowerShell-7.7.0-win-x64.zip','7c74e03b20a75b59715847439d046a6214b1863ee79b187872980f139bb4a05f'))){if((Get-FileHash $x[0]).Hash.ToLowerInvariant() -ne $x[1]){throw 'Archive hash mismatch'}}
Expand-Archive 'Microsoft.WindowsTerminal_1.25.3131.0_x64.zip' terminal
Expand-Archive 'PowerShell-7.7.0-win-x64.zip' pwsh7
$terminal=(Get-ChildItem terminal -Recurse -Filter WindowsTerminal.exe | Select-Object -First 1).FullName
$terminalRoot=Split-Path $terminal
New-Item -ItemType File -Path (Join-Path $terminalRoot '.portable') | Out-Null
New-Item -ItemType Directory -Path (Join-Path $terminalRoot 'settings') | Out-Null
foreach($d in @('desktop','desktop/commands','desktop/responses','desktop/done')){New-Item -ItemType Directory -Path $d | Out-Null}
$pwsh=(Resolve-Path 'pwsh7/pwsh.exe').Path
$manager='$b=Join-Path $env:USERPROFILE "asd-acceptance-20261122"; & "$b\product\asd-windows-amd64.exe" --backend "$b\backend\psmux.exe" --runtime-dir "$b\desktop-runtime" --namespace asd-desktop-20261122'
$encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($manager))
$settings=@{defaultProfile='{a81a7af0-f002-4a1e-982a-202611220001}';initialCols=120;initialRows=32;copyOnSelect=$false;profiles=@{list=@(@{guid='{a81a7af0-f002-4a1e-982a-202611220001}';name='ASD isolated desktop acceptance';commandline=('"'+$pwsh+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+$encoded);startingDirectory=$base;closeOnExit='never';font=@{face='Consolas';size=12};padding='0'})}}
$settings | ConvertTo-Json -Depth 8 | Set-Content (Join-Path $terminalRoot 'settings/settings.json') -Encoding UTF8
$asd=(Resolve-Path 'product/asd-windows-amd64.exe').Path
$common=@('--backend',(Resolve-Path 'backend/psmux.exe').Path,'--runtime-dir',(Join-Path $base 'desktop-runtime'),'--namespace','asd-desktop-20261122')
$initial=& $asd @common list --json | ConvertFrom-Json
if($LASTEXITCODE -ne 0 -or @($initial).Count -ne 0){throw 'Desktop namespace not empty or inaccessible'}
$cli='$ErrorActionPreference="Stop";[Console]::OutputEncoding=[Text.UTF8Encoding]::new();function Record($o){[IO.File]::AppendAllText($env:ASD_TEST_JOURNAL,($o|ConvertTo-Json -Compress)+"`n",[Text.UTF8Encoding]::new($false))};Record @{kind="started";pid=$PID;utc=[DateTime]::UtcNow.ToString("o")};Write-Output "ASD desktop fixture ready";while($null -ne ($line=[Console]::ReadLine())){Record @{kind="input";line=$line;utc=[DateTime]::UtcNow.ToString("o")};if($line -eq "exit"){break};Write-Output ("ASD received: "+$line)}'
$enc=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($cli))
$created=@()
foreach($entry in @(@('A','中文输入 A'),@('B','会话 B'),@('C','信号测试 C'))){
 $raw=& $asd @common start --name $entry[1] --env ('ASD_TEST_JOURNAL='+(Join-Path $base ('desktop/fixture-'+$entry[0]+'.jsonl'))) --cwd $base -- (Join-Path $PSHOME 'powershell.exe') -NoLogo -NoProfile -EncodedCommand $enc
 if($LASTEXITCODE -ne 0){throw 'Fixture create failed'};$created+=,($raw|ConvertFrom-Json);$created|ConvertTo-Json -Depth 6|Set-Content desktop/created.json -Encoding UTF8
}
$invoke='& ([ScriptBlock]::Create([IO.File]::ReadAllText((Join-Path $env:USERPROFILE "asd-acceptance-20261122\desktop-driver.ps1"))))'
$task='AsdDesktopAcceptance20261122'
if(Get-ScheduledTask -TaskName $task -ErrorAction SilentlyContinue){throw 'Refuse overwriting existing task'}
$action=New-ScheduledTaskAction -Execute powershell.exe -Argument ('-NoProfile -WindowStyle Hidden -STA -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($invoke)))
$principal=New-ScheduledTaskPrincipal -UserId ([Security.Principal.WindowsIdentity]::GetCurrent().User.Value) -LogonType Interactive -RunLevel Limited
$settings=New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Minutes 18)
Register-ScheduledTask -TaskName $task -Action $action -Principal $principal -Settings $settings|Out-Null
Start-ScheduledTask $task
@{terminal_version=(Get-Item $terminal).VersionInfo.FileVersion;pwsh_version=(& $pwsh -NoProfile -Command '$PSVersionTable.PSVersion.ToString()');created=$created;task=$task} | ConvertTo-Json -Depth 6
