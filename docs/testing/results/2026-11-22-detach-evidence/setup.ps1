$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new()
$b=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';Set-Location $b
if(Test-Path detach-diagnosis){throw 'Refuse existing diagnosis directory'}
$terminal=(Get-ChildItem terminal -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1).FullName
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object {$_.Path-eq$terminal}).Count){throw 'Private Terminal still running'}
foreach($d in @('detach-diagnosis','detach-diagnosis/commands','detach-diagnosis/responses','detach-diagnosis/done')){New-Item -ItemType Directory $d|Out-Null}
$common=@('--backend',"$b\backend\psmux.exe",'--runtime-dir',"$b\detach-runtime",'--namespace','asd-detach-20261122')
$asd="$b\product\asd-windows-amd64.exe"
$raw=& $asd @common start --name '分离诊断' --env ('ASD_TEST_JOURNAL='+"$b\detach-diagnosis\fixture.jsonl") --cwd $b -- "$b\desktop\native-reader.exe"
if($LASTEXITCODE-ne 0){throw 'Create diagnostic fixture failed'}
$raw-join "`n"|Set-Content detach-diagnosis/created.json -Encoding UTF8
$settingsPath=Join-Path (Split-Path $terminal) 'settings/settings.json'
Copy-Item $settingsPath detach-diagnosis/terminal-settings-before.json
$cfg=Get-Content $settingsPath -Raw|ConvertFrom-Json
$manager='$b=Join-Path $env:USERPROFILE "asd-acceptance-20261122"; & "$b\product\asd-windows-amd64.exe" --backend "$b\backend\psmux.exe" --runtime-dir "$b\detach-runtime" --namespace asd-detach-20261122'
$encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($manager))
$cfg.profiles.list[0].commandline=('"'+"$b\pwsh7\pwsh.exe"+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+$encoded)
$cfg|ConvertTo-Json -Depth 12|Set-Content $settingsPath -Encoding UTF8
$raw
