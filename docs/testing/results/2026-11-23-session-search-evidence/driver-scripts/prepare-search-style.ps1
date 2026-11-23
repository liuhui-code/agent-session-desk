$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'session-search-20261123'
$terminal=Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe'
if(@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in @($terminal,(Join-Path $base 'search-product.exe'),(Join-Path $base 'search-product-final.exe'))}).Count){throw 'Previous owned UI or window remains'}
if([string](Get-ScheduledTask AsdSessionSearchDesktop20261123).State -eq 'Running'){throw 'Previous driver still running'}
$settings=Join-Path (Split-Path $terminal) 'settings/settings.json';$receipt=[IO.File]::ReadAllText((Join-Path $root 'receipt.json'))|ConvertFrom-Json
if((Get-FileHash $settings).Hash -ne $receipt.settings_test){throw 'Settings changed externally'}
$product=Join-Path $base 'search-product-final.exe';if((Get-FileHash $product).Hash.ToLowerInvariant() -cne 'a7f5aa4cbcf64e728bdb0af84825c4e47df82b1bc3639ecd8b548c0f2fa58c4f'){throw 'Final product differs'}
$bundle=Join-Path $base 'terminal-bundle-20261123/bundle';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$command="& '$product' --terminal-bundle '$bundle' --runtime-dir '$(Join-Path $root 'runtime')' --namespace asd-session-search --state-dir '$(Join-Path $root 'state')'"
$config=[IO.File]::ReadAllText($settings)|ConvertFrom-Json;$profile=@($config.profiles.list|Where-Object{$_.guid -eq $config.defaultProfile})[0]
$profile.commandline='"'+$shell+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
[IO.File]::WriteAllText($settings,($config|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
@{utc=[DateTime]::UtcNow.ToString('o');settings_before=$receipt.settings_before;settings_before_style=$receipt.settings_test;settings_test=(Get-FileHash $settings).Hash;product_sha256=(Get-FileHash $product).Hash;scope='Final placeholder/font roles only; same three original shell Runs'}|ConvertTo-Json|Set-Content (Join-Path $root 'style-receipt.json') -Encoding utf8
Get-Content (Join-Path $root 'style-receipt.json') -Raw -Encoding utf8
