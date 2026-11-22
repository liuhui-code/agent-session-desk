$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$wsl=Get-Command wsl.exe -ErrorAction SilentlyContinue
$wslResult=$null;$wslCode=$null
if($wsl){$ErrorActionPreference='Continue';$wslResult=(& $wsl.Source --list --quiet 2>&1|Out-String);$wslCode=$LASTEXITCODE;$ErrorActionPreference='Stop'}
@{utc=[DateTime]::UtcNow.ToString('o');wsl_path=$wsl.Source;wsl_code=$wslCode;wsl_distributions=$wslResult;terminal_private_windows=@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -like ($base+'*')}).Count}|ConvertTo-Json -Depth 4
