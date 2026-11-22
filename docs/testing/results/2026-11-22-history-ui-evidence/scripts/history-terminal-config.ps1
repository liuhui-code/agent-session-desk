$ErrorActionPreference='Stop'
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'm1-history-ui-run2'
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
$path=Join-Path $terminal.Directory.FullName 'settings/settings.json'
$backup=Join-Path $root 'terminal-settings.before'
if(Test-Path $backup){throw 'Backup already exists'}
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object {$_.Path -eq $terminal.FullName}).Count -ne 0){throw 'Private terminal already running'}
[IO.File]::Copy($path,$backup)
$config=[IO.File]::ReadAllText($path)|ConvertFrom-Json
$code=[IO.File]::ReadAllText((Join-Path $root 'launch.txt'))
$encoded=[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($code))
$profile=@($config.profiles.list|Where-Object {$_.guid -eq $config.defaultProfile})[0]
$profile.commandline='"'+(Join-Path $base 'pwsh7/pwsh.exe')+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+$encoded
[IO.File]::WriteAllText($path,($config|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
@{settings_before_sha256=(Get-FileHash $backup).Hash;settings_test_sha256=(Get-FileHash $path).Hash;terminal_sha256=(Get-FileHash $terminal.FullName).Hash;command=$profile.commandline}|ConvertTo-Json|Set-Content (Join-Path $root 'terminal-config.json') -Encoding utf8
