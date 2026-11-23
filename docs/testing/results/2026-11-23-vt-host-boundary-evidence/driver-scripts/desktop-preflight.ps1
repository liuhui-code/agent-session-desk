$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'history-desktop-20261123'
if(Test-Path $root){throw 'Fresh desktop root required'}
$items=@(@('copy-final2-product.exe','a2518bc0c28c4ff23ac544b440fe1fcc704c1d0e682bbfaf3cb7da91d65934b0'),@('history-backend-candidate/psmux.exe','ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'),@('pwsh7/pwsh.exe','f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336'),@('terminal/terminal-1.25.3131.0/WindowsTerminal.exe','337c583db0b4b53c726a851933e3d3df86c09e0d4812d14d68ad703e81293fc9'))
$hashes=@{};foreach($i in $items){$p=Join-Path $base $i[0];$h=(Get-FileHash $p).Hash.ToLowerInvariant();if($h -ne $i[1]){throw 'Test binary differs'};$hashes[$i[0]]=$h}
$paths=@($items|ForEach-Object{Join-Path $base $_[0]})
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths -and $_.ExecutablePath -notlike '*pwsh7*'}|ForEach-Object{@{pid=$_.ProcessId;path=$_.ExecutablePath}})
if($live.Count){throw 'Exact test image already running'}
foreach($n in @('AsdHistoryDesktop20261123','AsdCopyFinal20261123')){if(Get-ScheduledTask $n -ErrorAction SilentlyContinue){throw 'Named desktop task already exists'}}
@{result='passed';hashes=$hashes;settings_sha256=(Get-FileHash (Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json')).Hash;private_images_alive=$live.Count;utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString}|ConvertTo-Json -Depth 5
