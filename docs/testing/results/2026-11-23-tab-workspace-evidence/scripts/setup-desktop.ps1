$ErrorActionPreference='Stop'
[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'tabs-ui-20261123'
$work=Join-Path $root '工作区 with spaces'
New-Item -ItemType Directory -Path $work -ErrorAction Stop|Out-Null
$product=Join-Path $base 'tabs-ui.exe';$psmux=Join-Path $base 'backend/psmux.exe'
$globalArgs=@('--backend',$psmux,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-tabs-ui','--state-dir',(Join-Path $root 'state'))
$created=@()
foreach($name in @('Shell A','Shell B')){
 $s=(& $product @globalArgs start --name $name --cwd $work -- (Join-Path $base 'pwsh7/pwsh.exe') -NoLogo -NoProfile |Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0){throw 'Shell start failed'}
 $created+=,$s
}
$created|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'sessions.json') -Encoding utf8
$code="& '$product' --backend '$psmux' --runtime-dir '$(Join-Path $root 'runtime')' --namespace asd-tabs-ui --state-dir '$(Join-Path $root 'state')'"
[IO.File]::WriteAllText((Join-Path $root 'launch.txt'),$code)
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName}).Count -ne 0){throw 'Private terminal already running'}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';$backup=Join-Path $root 'terminal-settings.before'
[IO.File]::Copy($settings,$backup)
$config=[IO.File]::ReadAllText($settings)|ConvertFrom-Json
$profile=@($config.profiles.list|Where-Object{$_.guid -eq $config.defaultProfile})[0]
$encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code))
$profile.commandline='"'+(Join-Path $base 'pwsh7/pwsh.exe')+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+$encoded
[IO.File]::WriteAllText($settings,($config|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
@{utc=[DateTime]::UtcNow.ToString('o');settings_before=(Get-FileHash $backup).Hash;settings_test=(Get-FileHash $settings).Hash;terminal_sha256=(Get-FileHash $terminal.FullName).Hash;sessions=$created}|ConvertTo-Json -Depth 8
