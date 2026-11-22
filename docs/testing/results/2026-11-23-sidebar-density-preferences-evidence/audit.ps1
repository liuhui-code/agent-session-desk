$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$terminal=Get-ChildItem (Join-Path $base 'terminal') -Filter WindowsTerminal.exe -Recurse|Select-Object -First 1
$files=@((Join-Path $base 'backend/psmux.exe'),(Join-Path $base 'pwsh7/pwsh.exe'),$terminal.FullName,(Join-Path $terminal.Directory.FullName 'OpenConsole.exe'))
$hashes=@();foreach($file in $files){$hashes+=,@{path=$file;sha256=(Get-FileHash $file).Hash}}
$settings=Join-Path $terminal.Directory.FullName 'settings/settings.json'
if((Get-FileHash $settings).Hash -ne '1499307ABEBA68C23F9EB38A4C636F81172475B6B2B535BC0B11DCD9547241F8'){throw 'Settings not restored'}
$checks=@()
foreach($rootName in @('density-20261123')){
 $flow=[IO.File]::ReadAllText((Join-Path $base "$rootName/flow-result.json"))|ConvertFrom-Json
 foreach($identity in $flow.shell_processes){
  $p=Get-Process -Id $identity.pid -ErrorAction SilentlyContinue
  $same=$p -and $p.StartTime.ToUniversalTime().ToFileTimeUtc() -eq $identity.filetime
  if($same){throw 'Original shell still alive after owned stop'}
  $checks+=,@{pid=$identity.pid;filetime=$identity.filetime;original_process_alive=$false}
 }
}
foreach($name in @('AsdDensity20261123')){if(Get-ScheduledTask -TaskName $name -ErrorAction SilentlyContinue){throw 'Owned task not removed'}}
$config=[IO.File]::ReadAllText($settings)|ConvertFrom-Json
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');binaries=$hashes;original_shells=$checks;tasks_removed=$true;settings_restored=(Get-FileHash $settings).Hash;font_defaults=$config.profiles.defaults.font;theme=$config.theme}|ConvertTo-Json -Depth 6
