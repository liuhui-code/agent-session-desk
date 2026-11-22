$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'responsive-20261123'
New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null
$product=Join-Path $base 'responsive-ui.exe';$backend=Join-Path $base 'backend/psmux.exe';$test=Join-Path $base 'responsive-ui-tests.exe'
& $test '-test.v' *> (Join-Path $root 'ui-tests.log');if($LASTEXITCODE -ne 0){throw 'Native UI tests failed'}
$g=@('--backend',$backend,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-responsive','--state-dir',(Join-Path $root 'state'))
$work=Join-Path $root 'work';New-Item -ItemType Directory -Path $work|Out-Null
$sessions=@()
foreach($name in @('Responsive Shell')){
 $text=(& $product @g start --name $name --cwd $work -- (Join-Path $base 'pwsh7/pwsh.exe') -NoLogo -NoProfile|Out-String)
 if($LASTEXITCODE -ne 0){throw 'Shell start failed'}
 $sessions+=,($text|ConvertFrom-Json)
}
$sessions|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'sessions.json') -Encoding utf8
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName}).Count -ne 0){throw 'Private terminal already running'}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';$backup=Join-Path $root 'settings.before';[IO.File]::Copy($settings,$backup)
$config=[IO.File]::ReadAllText($settings)|ConvertFrom-Json;$profile=@($config.profiles.list|Where-Object{$_.guid -eq $config.defaultProfile})[0]
$command="& '$product' --backend '$backend' --runtime-dir '$(Join-Path $root 'runtime')' --namespace asd-responsive --state-dir '$(Join-Path $root 'state')'"
$profile.commandline='"'+(Join-Path $base 'pwsh7/pwsh.exe')+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
[IO.File]::WriteAllText($settings,($config|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
@{utc=[DateTime]::UtcNow.ToString('o');app_sha256=(Get-FileHash $product).Hash;test_sha256=(Get-FileHash $test).Hash;backend_sha256=(Get-FileHash $backend).Hash;settings_before=(Get-FileHash $backup).Hash;settings_test=(Get-FileHash $settings).Hash;windows=[Environment]::OSVersion.VersionString}|ConvertTo-Json|Set-Content (Join-Path $root 'receipt.json') -Encoding utf8
Get-Content (Join-Path $root 'receipt.json') -Raw -Encoding utf8
