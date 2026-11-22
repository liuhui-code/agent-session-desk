$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'visual-replay-utf8-20261123'
New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName}).Count -ne 0){throw 'Private terminal already running'}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';$backup=Join-Path $root 'settings.before';[IO.File]::Copy($settings,$backup)
$config=[IO.File]::ReadAllText($settings)|ConvertFrom-Json;$profile=@($config.profiles.list|Where-Object{$_.guid -eq $config.defaultProfile})[0]
$command="& '$(Join-Path $base 'replay-reference-utf8.ps1')'"
$profile.commandline='"'+(Join-Path $base 'pwsh7/pwsh.exe')+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
foreach($pair in @(@('font',@{face='Consolas';size=12}),@('background','#1a1b26'),@('foreground','#c0caf5'),@('cursorColor','#c0caf5'))){$profile|Add-Member -MemberType NoteProperty -Name $pair[0] -Value $pair[1] -Force}
[IO.File]::WriteAllText($settings,($config|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
@{font=$profile.font;background=$profile.background;foreground=$profile.foreground;cursorColor=$profile.cursorColor;colorScheme=$profile.colorScheme}|ConvertTo-Json -Depth 5|Set-Content (Join-Path $root 'test-display-profile.json') -Encoding utf8
@{utc=[DateTime]::UtcNow.ToString('o');settings_before=(Get-FileHash $backup).Hash;settings_test=(Get-FileHash $settings).Hash}|ConvertTo-Json|Set-Content (Join-Path $root 'receipt.json') -Encoding utf8
Get-Content (Join-Path $root 'receipt.json') -Raw -Encoding utf8
