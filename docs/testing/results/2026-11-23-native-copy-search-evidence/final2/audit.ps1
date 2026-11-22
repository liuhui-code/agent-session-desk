$ErrorActionPreference='Stop';$ProgressPreference='SilentlyContinue';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122';$root=Join-Path $base 'copy-final-20261123'
$flow=[IO.File]::ReadAllText((Join-Path $root 'flow-result.json'))|ConvertFrom-Json
$fixtures=@($flow.shell_processes);if($fixtures.Count -ne 2){throw 'Expected exactly two recorded shell identities'}
$ui=[IO.File]::ReadAllText((Join-Path $root 'replaced-ui-process.json'))|ConvertFrom-Json;$identities=@($fixtures)+@($ui);$ended=@()
foreach($id in $identities){$p=Get-Process -Id $id.pid -ErrorAction SilentlyContinue;$same=$false;if($p){$same=$p.Path -eq $id.path -and $p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $id.filetime};if($same){throw 'Original owned process still alive'};$ended+=,@{pid=$id.pid;filetime=$id.filetime;path=$id.path;original_ended=$true;pid_reused=[bool]$p}}
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
$paths=@((Join-Path $base 'copy-final2-product.exe'),(Join-Path $base 'copy-final2-ui-tests.exe'),(Join-Path $base 'copy-final2-terminal-tests.exe'),(Join-Path $base 'copy-final-product.exe'),$terminal.FullName)
$live=@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths});if($live.Count -ne 0){throw 'Exact test image remains alive'}
foreach($name in @('AsdCopyFinal20261123')){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){throw 'Owned task remains'}}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json';$hash=(Get-FileHash $settings).Hash
$clipboard=[IO.File]::ReadAllText((Join-Path $root 'desktop/clipboard-restored.json'))|ConvertFrom-Json;if(!$clipboard.restored){throw 'Clipboard restore mismatch'}
if($hash -ne '1499307ABEBA68C23F9EB38A4C636F81172475B6B2B535BC0B11DCD9547241F8'){throw 'Original settings not restored'}
$bins=@();foreach($path in @($paths)+@((Join-Path $base 'backend/psmux.exe'),(Join-Path $base 'pwsh7/pwsh.exe'),(Join-Path $terminal.Directory.FullName 'OpenConsole.exe'))){$bins+=,@{path=$path;sha256=(Get-FileHash $path).Hash}}
@{clipboard_restored=$clipboard.restored;result='passed';utc=[DateTime]::UtcNow.ToString('o');original_processes=$ended;exact_test_images_alive=0;tasks_absent=$true;settings_restored=$hash;binaries=$bins}|ConvertTo-Json -Depth 6|Set-Content (Join-Path $root 'final-audit.json') -Encoding utf8
Get-Content (Join-Path $root 'final-audit.json') -Raw -Encoding utf8
