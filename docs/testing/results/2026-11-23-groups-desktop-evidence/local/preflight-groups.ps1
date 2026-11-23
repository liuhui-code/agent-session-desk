$ErrorActionPreference='Stop';[Console]::OutputEncoding=[Text.UTF8Encoding]::new($false)
$base=Join-Path $env:USERPROFILE 'asd-acceptance-20261122'
$root=Join-Path $base 'group-workspace-desktop-20261123'
if(Test-Path $root){throw 'Fresh desktop root required'}
$expected=@(@('asd-group-desktop.exe','d8b88e66425fe284386aa5dd3a10d4fd0efadad490e31721d9412f01a1b2a6ed'),@('read-group-metadata.exe','15d1395962ec77a44f0f7768acc83461f47804b0d2f4fd7711b3b59800c746ce'),@('terminal-bundle-20261123/bundle/psmux.exe','ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937'),@('terminal-bundle-20261123/bundle/conpty.dll','e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492'),@('terminal-bundle-20261123/bundle/OpenConsole.exe','b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775'),@('pwsh7/pwsh.exe','f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336'),@('terminal/terminal-1.25.3131.0/WindowsTerminal.exe','337c583db0b4b53c726a851933e3d3df86c09e0d4812d14d68ad703e81293fc9'))
foreach($e in $expected){if((Get-FileHash (Join-Path $base $e[0])).Hash.ToLowerInvariant() -cne $e[1]){throw ('Pinned binary differs: '+$e[0])}}
$paths=@('asd-group-desktop.exe','terminal-bundle-20261123/bundle/psmux.exe','terminal-bundle-20261123/bundle/OpenConsole.exe','terminal/terminal-1.25.3131.0/WindowsTerminal.exe')|ForEach-Object{Join-Path $base $_}
if(@(Get-CimInstance Win32_Process|Where-Object{$_.ExecutablePath -in $paths}).Count){throw 'Known private image already live'}
foreach($name in @('AsdGroupWorkspaceSetup20261123','AsdGroupWorkspaceDesktop20261123')){if(Get-ScheduledTask $name -ErrorAction SilentlyContinue){throw 'Owned task exists'}}
$settings=Join-Path $base 'terminal/terminal-1.25.3131.0/settings/settings.json'
if((Get-FileHash $settings).Hash.ToLowerInvariant() -cne '1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8'){throw 'Dedicated settings changed'}
foreach($file in @('setup-groups-desktop.ps1','start-groups-setup.ps1','start-groups-desktop.ps1','groups-desktop-driver.ps1','groups-flow.ps1','cleanup-export-groups.ps1')){
 $tokens=$null;$errors=$null;$null=[Management.Automation.Language.Parser]::ParseInput([IO.File]::ReadAllText((Join-Path $base $file),[Text.Encoding]::UTF8),[ref]$tokens,[ref]$errors)
 if($errors.Count){throw ('Script syntax invalid: '+$file+' '+$errors[0].Message)}
}
@{result='passed';utc=[DateTime]::UtcNow.ToString('o');windows=[Environment]::OSVersion.VersionString;known_private_images_alive=0;fresh_root=$true;settings_sha256=(Get-FileHash $settings).Hash;hashes_checked=$expected.Count;parsed_scripts=6}|ConvertTo-Json
