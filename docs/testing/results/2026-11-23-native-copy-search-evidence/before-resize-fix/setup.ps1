$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'copy-final-20261123'
New-Item -ItemType Directory -Path $root -ErrorAction Stop|Out-Null
$product=Join-Path $base 'copy-final-product.exe';$backend=Join-Path $base 'backend/psmux.exe';$test=Join-Path $base 'copy-final-ui-tests.exe'
& $test '-test.v' *> (Join-Path $root 'ui-tests.log');if($LASTEXITCODE -ne 0){throw 'Native UI tests failed'}
$terminalTest=Join-Path $base 'copy-final-terminal-tests.exe'
& $terminalTest '-test.v' *> (Join-Path $root 'terminal-tests.log');if($LASTEXITCODE -ne 0){throw 'Native terminal tests failed'}
$g=@('--backend',$backend,'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-copy-final','--state-dir',(Join-Path $root 'state'))
$work=Join-Path $root 'work';New-Item -ItemType Directory -Path $work|Out-Null
$promptScript=Join-Path $work 'prompt.ps1'
$promptText=@'
param([string]$label)
function global:prompt { '> ' }
Clear-Host
'SCROLL_NEEDLE_中文_α'
1..40 | ForEach-Object { 'history line '+$_ }
Write-Output ('COPY_READY '+$label)
'@
[IO.File]::WriteAllText($promptScript,$promptText,[Text.UTF8Encoding]::new($false))
$sessions=@()
foreach($name in @('Copy A','Copy B')){
 $text=(& $product @g start --name $name --cwd $work -- (Join-Path $base 'pwsh7/pwsh.exe') -NoLogo -NoProfile -NoExit -File $promptScript $name|Out-String)
 if($LASTEXITCODE -ne 0){throw 'Shell start failed'}
 $sessions+=,($text|ConvertFrom-Json)
}
$sessions|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'sessions.json') -Encoding utf8
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName}).Count -ne 0){throw 'Private terminal already running'}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';$backup=Join-Path $root 'settings.before';[IO.File]::Copy($settings,$backup)
$config=[IO.File]::ReadAllText($settings)|ConvertFrom-Json;$profile=@($config.profiles.list|Where-Object{$_.guid -eq $config.defaultProfile})[0]
$command="& '$product' --backend '$backend' --runtime-dir '$(Join-Path $root 'runtime')' --namespace asd-copy-final --state-dir '$(Join-Path $root 'state')'"
$profile.commandline='"'+(Join-Path $base 'pwsh7/pwsh.exe')+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
[IO.File]::WriteAllText($settings,($config|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
@{utc=[DateTime]::UtcNow.ToString('o');app_sha256=(Get-FileHash $product).Hash;test_sha256=(Get-FileHash $test).Hash;terminal_test_sha256=(Get-FileHash $terminalTest).Hash;backend_sha256=(Get-FileHash $backend).Hash;settings_before=(Get-FileHash $backup).Hash;settings_test=(Get-FileHash $settings).Hash;windows=[Environment]::OSVersion.VersionString}|ConvertTo-Json|Set-Content (Join-Path $root 'receipt.json') -Encoding utf8
Get-Content (Join-Path $root 'receipt.json') -Raw -Encoding utf8
