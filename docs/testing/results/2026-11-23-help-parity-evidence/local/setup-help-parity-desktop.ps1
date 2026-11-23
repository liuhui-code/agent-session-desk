$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'help-parity-desktop-20261123'
if(Test-Path $root){throw 'Fresh root required'}
$product=Join-Path $base 'asd-help-parity.exe';$backend=Join-Path $base 'terminal-bundle-20261123/bundle/psmux.exe';$shell=Join-Path $base 'pwsh7/pwsh.exe'
$terminal=Get-Item (Join-Path $base 'terminal/terminal-1.25.3131.0/WindowsTerminal.exe')
$expected=@(@($product,'b84f1d8b08e064423a341c923af5079589c4f0384ffc5e008b7bf32b4980056d'),@($backend,'ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'),@($shell,'f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336'),@($terminal.FullName,'337c583db0b4b53c726a851933e3d3df86c09e0d4812d14d68ad703e81293fc9'))
$expected+=,@((Join-Path (Split-Path $backend) 'conpty.dll'),'e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492')
$expected+=,@((Join-Path (Split-Path $backend) 'OpenConsole.exe'),'b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775')
foreach($v in $expected){if((Get-FileHash $v[0]).Hash.ToLowerInvariant() -ne $v[1]){throw 'Binary differs'}}
if(@(Get-Process WindowsTerminal -ErrorAction SilentlyContinue|Where-Object{$_.Path -eq $terminal.FullName}).Count){throw 'Private terminal already running'}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json'
if((Get-FileHash $settings).Hash.ToLowerInvariant() -ne '1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'){throw 'Settings changed since preflight'}
New-Item -ItemType Directory $root|Out-Null;$work=Join-Path $root 'work';New-Item -ItemType Directory $work|Out-Null
$g=@('--terminal-bundle',(Split-Path $backend),'--runtime-dir',(Join-Path $root 'runtime'),'--namespace','asd-help-parity','--state-dir',(Join-Path $root 'state'))
foreach($dir in @('root-default','review-default')){New-Item -ItemType Directory (Join-Path $work $dir)|Out-Null}
$promptScript=Join-Path $work 'prompt.ps1'
@'
param([string]$label)
$identityRoot=Split-Path -Parent $PSCommandPath
function global:prompt { '> ' }
Set-PSReadLineOption -HistorySavePath (Join-Path $identityRoot ($label.Replace(' ','_')+'.history.txt'))
$p=Get-Process -Id $PID
@{label=$label;pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path;cwd=$PWD.Path}|ConvertTo-Json|Set-Content (Join-Path $identityRoot ($label.Replace(' ','_')+'.identity.json')) -Encoding utf8
Clear-Host
'SCROLL_NEEDLE_中文_α'
1..40 | ForEach-Object { 'history line '+$_ }
Write-Output ('COPY_READY '+$label)
'@|Set-Content $promptScript -Encoding utf8
@'
$ErrorActionPreference='Stop'
$p=Get-Process -Id $PID
@{pid=$PID;filetime=$p.StartTime.ToUniversalTime().ToFileTimeUtc();path=$p.Path}|ConvertTo-Json|Set-Content alt-started.json -Encoding utf8
[Console]::Write("`e[?1049h`e[2J`e[HVT_ALT_READY 中文 α")
try{
 $line=[Console]::ReadLine()
 $cols=[Console]::WindowWidth;$rows=[Console]::WindowHeight
 @{pid=$PID;text=$line;columns=$cols;rows=$rows}|ConvertTo-Json|Set-Content alt-observed.json -Encoding utf8
}finally{[Console]::Write("`e[?1049l")}
'VT_ALT_DONE'
'@|Set-Content (Join-Path $work 'alt-app.ps1') -Encoding utf8
$sessions=@()
foreach($name in @('Copy A','Copy B')){
 $s=(& $product @g start --name $name --cwd $work -- $shell -NoLogo -NoProfile -NoExit -File $promptScript $name|Out-String)|ConvertFrom-Json
 if($LASTEXITCODE -ne 0){throw 'Shell start failed'};$sessions+=,$s
 $sessions|ConvertTo-Json -Depth 8|Set-Content (Join-Path $root 'sessions.json') -Encoding utf8
}
$until=[DateTime]::UtcNow.AddSeconds(10)
foreach($name in @('Copy_A','Copy_B')){
 $file=Join-Path $work ($name+'.identity.json');while(!(Test-Path $file) -and [DateTime]::UtcNow -lt $until){Start-Sleep -Milliseconds 50}
 $id=[IO.File]::ReadAllText($file)|ConvertFrom-Json;$p=Get-Process -Id $id.pid
 if($p.Path -ne $shell -or $p.StartTime.ToUniversalTime().ToFileTimeUtc() -ne $id.filetime){throw 'Original shell identity mismatch'}
}
$ui=Join-Path $root 'state/ui.toml'
[IO.File]::WriteAllText($ui,"schema_version = 3`nsidebar_density = `"compact`"`nsidebar_mode = `"grouped`"`ndefault_group_collapsed = false`ncollapsed_groups = `"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee`"`n",[Text.UTF8Encoding]::new($false))
[IO.File]::Copy($ui,(Join-Path $root 'ui-v3-before.toml'))
$backup=Join-Path $root 'settings.before' ;[IO.File]::Copy($settings,$backup)
$config=[IO.File]::ReadAllText($settings)|ConvertFrom-Json;$profile=@($config.profiles.list|Where-Object{$_.guid -eq $config.defaultProfile})[0]
$command="Import-Module PSReadLine; Set-PSReadLineOption -HistorySavePath '$(Join-Path $work 'ui-history.txt')'; & '$product' --terminal-bundle '$(Split-Path $backend)' --runtime-dir '$(Join-Path $root 'runtime')' --namespace asd-help-parity --state-dir '$(Join-Path $root 'state')'"
$profile.commandline='"'+$shell+'" -NoLogo -NoProfile -NoExit -EncodedCommand '+[Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($command))
[IO.File]::WriteAllText($settings,($config|ConvertTo-Json -Depth 12),[Text.UTF8Encoding]::new($false))
@{utc=[DateTime]::UtcNow.ToString('o');app_sha256=(Get-FileHash $product).Hash;backend_sha256=(Get-FileHash $backend).Hash;settings_before=(Get-FileHash $backup).Hash;settings_test=(Get-FileHash $settings).Hash;windows=[Environment]::OSVersion.VersionString;product_reused=$false;go_tests_rerun=$false;desktop_only=$true}|ConvertTo-Json|Set-Content (Join-Path $root 'receipt.json') -Encoding utf8
Get-Content (Join-Path $root 'receipt.json') -Raw -Encoding utf8
