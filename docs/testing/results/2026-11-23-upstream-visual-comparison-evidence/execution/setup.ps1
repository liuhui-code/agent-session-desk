$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'visual-compare-20261123'
New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null
$product=Join-Path $base 'product-after.exe';$backend=Join-Path $base 'backend/psmux.exe';$test=Join-Path $base 'compare-ui-tests.exe'
& $test '-test.v' *> (Join-Path $root 'ui-tests.log');if($LASTEXITCODE -ne 0){throw 'Native UI tests failed'}
$g=@('--backend',$backend,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-visual-compare','--state-dir',(Join-Path $root 'state'))
$work=Join-Path $root 'work';New-Item -ItemType Directory -Path $work|Out-Null
$sessions=@()
foreach($name in @('Review A','Review B')){
 $text=(& $product @g start --name $name --cwd $work -- (Join-Path $base 'visual-fixture.exe') $name|Out-String)
 if($LASTEXITCODE -ne 0){throw 'Shell start failed'}
 $sessions+=,($text|ConvertFrom-Json)
}
$sessions|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'sessions.json') -Encoding utf8
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName}).Count -ne 0){throw 'Private terminal already running'}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';$backup=Join-Path $root 'settings.before';[IO.File]::Copy($settings,$backup)
$config=[IO.File]::ReadAllText($settings)|ConvertFrom-Json;$profile=@($config.profiles.list|Where-Object{$_.guid -eq $config.defaultProfile})[0]
$command="& '$(Join-Path $base 'replay-reference.ps1')'"
$profile.commandline='"'+(Join-Path $base 'pwsh7/pwsh.exe')+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
foreach($pair in @(@('font',@{face='Consolas';size=12}),@('background','#1a1b26'),@('foreground','#c0caf5'),@('cursorColor','#c0caf5'))){$profile|Add-Member -MemberType NoteProperty -Name $pair[0] -Value $pair[1] -Force}
[IO.File]::WriteAllText($settings,($config|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
@{font=$profile.font;background=$profile.background;foreground=$profile.foreground;cursorColor=$profile.cursorColor;colorScheme=$profile.colorScheme}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'test-display-profile.json') -Encoding utf8
@{utc=[DateTime]::UtcNow.ToString('o');app_sha256=(Get-FileHash $product).Hash;test_sha256=(Get-FileHash $test).Hash;before_app_sha256=(Get-FileHash (Join-Path $base 'product-before.exe')).Hash;fixture_sha256=(Get-FileHash (Join-Path $base 'visual-fixture.exe')).Hash;backend_sha256=(Get-FileHash $backend).Hash;settings_before=(Get-FileHash $backup).Hash;settings_test=(Get-FileHash $settings).Hash;windows=[Environment]::OSVersion.VersionString}|ConvertTo-Json|Set-Content (Join-Path $root 'receipt.json') -Encoding utf8
Get-Content (Join-Path $root 'receipt.json') -Raw -Encoding utf8
